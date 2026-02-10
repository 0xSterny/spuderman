package matcher

var DefaultExcludes = []string{
	// OS / System
	`Windows`,
	`Program Files`,
	`Program Files (x86)`,
	`AppData`,
	`$Recycle.Bin`,
	`System Volume Information`,
	`.DS_Store`,
	`Thumbs.db`,

	// Dev / SCM
	`.git`,
	`.svn`,
	`node_modules`,
	`vendor`,
	`__pycache__`,
	`.idea`,
	`.vscode`,
	`venv`,
	// .env might contain secrets! User might want this.
	// Removing .env from excludes.

	// Media / Binary large
	`.iso`,
	`.exe`,
	`.dll`,
	`.bin`,
	`.msi`,

	// User-specific / Legacy Windows
	`Microsoft/Edge/User Data/ZxcvbnData`,
	`Windows/ServiceProfiles`,
	`Cookies`,
	`My Music`,
	`My Pictures`,
	`My Videos`,
	`NetHood`,
	`PrintHood`,
	`Recent`,
	`SendTo`,
	`Start Menu`,
	`Templates`,
	`Program Files`,
	`Program Files (x86)`,
	`ProgramData`,
	`Windows`,
	`System Volume Information`,
	`$Recycle.Bin`,
	`pagefile.sys`,
	`swapfile.sys`,
	`hiberfil.sys`,
	// Add more default exclusions here
}
