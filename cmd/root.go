package cmd

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/hirochachacha/go-smb2"
	"github.com/spf13/cobra"

	"spuderman/pkg/matcher"
	"spuderman/pkg/smbclient"
	"spuderman/pkg/spider"
	"spuderman/pkg/state"
	"spuderman/pkg/utils"
)

var (
	// Authentication
	username string
	password string
	domain   string

	hash      string
	ccache    string
	krbConfig string

	// Filters
	filenames  []string
	extensions []string
	content    []string
	sharenames []string
	dirnames   []string

	// Settings
	threads         int
	concurrentHosts int
	maxDepth        int
	analyze         bool
	lootDir         string
	noDownload      bool
	structuredLoot  bool
	verbose         bool
	noPass          bool

	// Reporting
	outputFile string

	// Phase 2
	presets   []string
	noExclude bool

	// Phase 3
	resumeFile string
)

var rootCmd = &cobra.Command{
	Use:   "spuderman [targets]",
	Short: "Spuderman: Spider entire networks for juicy files on SMB shares",
	Long: `Spuderman is a Go port of MANSPIDER. 
It spiders SMB shares (and local paths) to find sensitive files 
based on filenames, extensions, and content.

Targets can be:
- Single IP or Hostname (e.g. 192.168.1.1)
- CIDR Range (e.g. 192.168.1.0/24)
- File containing targets (one per line)
- Local Directory`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}

		// Initialize Logger
		if err := utils.InitLogger("spuderman.log"); err != nil {
			fmt.Printf("Failed to init logger: %v\n", err)
		}
		defer utils.CloseLogger()

		utils.PrintBanner() // Added Banner
		utils.LogInfo("Spuderman starting...")

		if analyze {
			utils.LogInfo("Analyze mode enabled: Disabling downloads, enabling verbose")
			noDownload = true
			verbose = true
			os.Setenv("DEBUG", "true")
		}

		if noPass {
			if verbose {
				utils.LogDebug("Using --no-pass, ignoring any provided password")
			}
			password = ""
		}

		// Global Deduplicator
		dedup := utils.NewDeduplicator()

		// Setup Reporter
		var reporter spider.Reporter
		if outputFile != "" {
			jr, err := spider.NewJSONReporter(outputFile)
			if err != nil {
				utils.LogError("Failed to create reporter: %v", err)
				return
			}
			defer jr.Close()
			reporter = jr
		} else {
			reporter = &spider.ConsoleReporter{} // Dummy
		}

		// 1. Setup Matcher
		mConfig := matcher.MatchConfig{
			Filenames:  filenames,
			Extensions: extensions,
			Content:    content,
			Dirnames:   dirnames,
			Presets:    presets, // Load presets
			OrLogic:    true,
		}
		matchEngine, err := matcher.NewMatcher(mConfig)
		if err != nil {
			utils.LogError("Failed to create matcher: %v", err)
			return
		}

		// Handle Default Exclusions
		if noExclude {
			utils.LogWarning("Disabling default exclusions")
			matchEngine.ExcludeRegex = nil // Clear defaults
		}

		// 2. Setup Spider Config
		sConfig := spider.Config{
			MaxDepth:   maxDepth,
			Threads:    threads,
			LootDir:    lootDir,
			NoDownload: noDownload,
			Structured: structuredLoot,
		}

		// Resume State
		var stateMgr *state.Manager
		if resumeFile != "" {
			var err error
			stateMgr, err = state.NewManager(resumeFile)
			if err != nil {
				utils.LogError("Failed to init state manager: %v", err)
				return
			}
			utils.LogInfo("Resume mode enabled. Loaded state from %s", resumeFile)
		}

		// 3. Process Targets
		var finalTargets []string

		for _, arg := range args {
			// A. Try CIDR
			_, ipnet, err := net.ParseCIDR(arg)
			if err == nil {
				// Expand CIDR
				utils.LogInfo("Expanding CIDR: %s", arg)
				for ip := ipnet.IP.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
					ipStr := ip.String()
					if stateMgr != nil && stateMgr.IsCompleted(ipStr) {
						continue
					}
					finalTargets = append(finalTargets, ipStr)
				}
				continue
			}

			// B. Check File
			fi, err := os.Stat(arg)
			if err == nil && !fi.IsDir() {
				utils.LogInfo("Reading targets from file: %s", arg)
				file, err := os.Open(arg)
				if err != nil {
					utils.LogError("Failed to open target file %s: %v", arg, err)
					continue
				}
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					t := strings.TrimSpace(scanner.Text())
					if t != "" {
						if stateMgr != nil && stateMgr.IsCompleted(t) {
							continue
						}
						finalTargets = append(finalTargets, t)
					}
				}
				file.Close()
				continue
			}

			// C. Default
			if stateMgr != nil && stateMgr.IsCompleted(arg) {
				utils.LogInfo("Skipping completed target: %s", arg)
				continue
			}
			finalTargets = append(finalTargets, arg)
		}

		if stateMgr != nil {
			utils.LogInfo("Targets after resume filter: %d", len(finalTargets))
		} else {
			utils.LogInfo("Total targets processed: %d", len(finalTargets))
		}

		var targetWG sync.WaitGroup
		targetSem := make(chan struct{}, concurrentHosts)

		var completedTargets int32
		totalTargets := len(finalTargets)

		for _, target := range finalTargets {
			targetWG.Add(1)
			targetSem <- struct{}{}

			go func(tgt string) {
				defer targetWG.Done()
				defer func() {
					<-targetSem
					curr := atomic.AddInt32(&completedTargets, 1)
					utils.LogInfo("Progress: Hosts [%d/%d] - Finished %s", curr, totalTargets, tgt)
					if stateMgr != nil {
						stateMgr.MarkCompleted(tgt)
					}
				}()

				// Check if local
				if _, err := os.Stat(tgt); err == nil {
					// Local path
					utils.LogInfo("Scanning local path: %s", tgt)
					fs := &spider.LocalFS{}

					localCfg := sConfig
					localCfg.Host = "Local"
					localCfg.Share = tgt
					s := spider.NewSpider(localCfg, matchEngine, fs, dedup, reporter)
					// Local scan uses own threads logic unless we want to bound it?
					// NewSpider defaults to creating its own sem if nil.
					// Since local is 1 "Host", it's fine.
					s.Walk(tgt)
				} else {
					// Assume SMB
					utils.LogInfo("Scanning remote target: %s", tgt)

					// Connect SMB
					session, err := smbclient.NewSession(tgt, username, password, domain, hash, ccache, krbConfig)
					if err != nil {
						utils.LogError("Failed to connect to %s: %v", tgt, err)
						return
					}
					defer session.Close()

					// List Shares (or use provided)
					var shares []string
					if len(sharenames) > 0 {
						shares = sharenames
					} else {
						var err error
						shares, err = session.ListShares()
						if err != nil {
							utils.LogError("Failed to list shares on %s: %v", tgt, err)
							if strings.Contains(err.Error(), "signing required") {
								utils.LogWarning("Target requires SMB Signing which interfered with Share Listing.")
								utils.LogWarning("TRY: specifying shares manually with --sharenames (e.g. '--sharenames C$,ADMIN$,Users')")
							}
							return
						}
					}

					// Shared Semaphore for this Host
					hostSem := make(chan struct{}, threads)
					var shareWG sync.WaitGroup

					for _, share := range shares {
						// Exclude IPC$ share (common request to avoid pipe scanning)
						if share == "IPC$" {
							if verbose {
								utils.LogDebug("Skipping IPC$ share")
							}
							continue
						}

						// Mount (Serial mounting is safer)
						mountedShare, err := session.Mount(share)
						if err != nil {
							utils.LogWarning("Failed to mount %s on %s: %v", share, tgt, err)
							continue
						}

						// Launch Share Scan
						shareWG.Add(1)
						go func(sh string, mount *smb2.Share) {
							defer shareWG.Done()
							// defer mount.Umount()? Configured in session.

							utils.LogInfo("Scanning share: \\\\%s\\%s", tgt, sh)
							fs := &spider.SMBFS{Share: mount}

							shareCfg := sConfig
							shareCfg.Host = tgt
							shareCfg.Share = sh

							s := spider.NewSpider(shareCfg, matchEngine, fs, dedup, reporter)
							s.Semaphore = hostSem // Inject shared semaphore
							s.Walk(".")
						}(share, mountedShare)
					}
					shareWG.Wait()
				}
			}(target)
		}
		targetWG.Wait()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Auth
	rootCmd.PersistentFlags().StringVarP(&username, "username", "u", "", "Username for authentication")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Password for authentication")
	rootCmd.PersistentFlags().StringVarP(&domain, "domain", "d", "", "Domain for authentication")
	rootCmd.PersistentFlags().StringVarP(&hash, "hash", "H", "", "NTLM hash for authentication")
	rootCmd.PersistentFlags().StringVarP(&ccache, "ccache", "", "", "Kerberos CCache file path")
	rootCmd.PersistentFlags().StringVarP(&krbConfig, "krb5-conf", "", "", "Kerberos config file path (krb5.conf)")
	rootCmd.PersistentFlags().BoolVarP(&noPass, "no-pass", "", false, "Do not use a password (force empty)")

	// Filters
	rootCmd.PersistentFlags().StringSliceVarP(&filenames, "filenames", "f", []string{}, "Filter filenames using regex")
	rootCmd.PersistentFlags().StringSliceVarP(&extensions, "extensions", "e", []string{}, "Only show filenames with these extensions")
	rootCmd.PersistentFlags().StringSliceVarP(&content, "content", "c", []string{}, "Search for file content using regex")
	rootCmd.PersistentFlags().StringSliceVar(&sharenames, "sharenames", []string{}, "Only search shares with these names")
	rootCmd.PersistentFlags().StringSliceVar(&dirnames, "dirnames", []string{}, "Only search directories containing these strings")

	// Config
	rootCmd.PersistentFlags().IntVarP(&threads, "threads", "t", 5, "Concurrent threads (PER HOST)")
	rootCmd.PersistentFlags().IntVarP(&concurrentHosts, "parallel", "P", 5, "Max concurrent hosts")
	rootCmd.PersistentFlags().IntVarP(&maxDepth, "maxdepth", "m", 10, "Maximum depth to spider")
	rootCmd.PersistentFlags().BoolVarP(&analyze, "analyze", "A", false, "Analyze mode: No download, Verbose output, Log to file")
	rootCmd.PersistentFlags().StringVarP(&lootDir, "loot-dir", "l", ".spuderman/loot", "Loot directory")
	rootCmd.PersistentFlags().BoolVarP(&structuredLoot, "structured", "S", false, "Use structured loot directory (Host/Share/File)")
	rootCmd.PersistentFlags().BoolVarP(&noDownload, "no-download", "n", false, "Don't download matching files")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show debugging messages")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "Output file for results (JSON)")

	// Phase 2: UX
	rootCmd.PersistentFlags().StringSliceVar(&presets, "preset", []string{}, "Load secret regex presets (e.g. aws, azure, slack, keys)")
	rootCmd.PersistentFlags().BoolVar(&noExclude, "no-exclude", false, "Disable default exclusions")

	// Phase 3
	rootCmd.PersistentFlags().StringVar(&resumeFile, "resume", "", "Resume state file (JSON)")
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
