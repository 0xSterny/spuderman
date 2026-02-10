package matcher

var SecretPresets = map[string][]string{
	"aws": {
		`(?i)aws_access_key_id`,
		`(?i)aws_secret_access_key`,
		`(?i)(A3T[A-Z0-9]|AKIA|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA)[A-Z0-9]{16}`,
	},
	"azure": {
		`(?i)AccountKey=[a-zA-Z0-9+/=]{88}`,
		`(?i)azure.*?key`,
	},
	"slack": {
		`xox[baprs]-([0-9a-zA-Z]{10,48})?`,
		`https://hooks.slack.com/services/T[a-zA-Z0-9_]{8}/B[a-zA-Z0-9_]{8}/[a-zA-Z0-9_]{24}`,
	},
	"keys": {
		`(?i)-----BEGIN RSA PRIVATE KEY-----`,
		`(?i)-----BEGIN OPENSSH PRIVATE KEY-----`,
		`(?i)-----BEGIN PRIVATE KEY-----`,
		`(?i)-----BEGIN EC PRIVATE KEY-----`,
		`(?i)-----BEGIN PGP PRIVATE KEY BLOCK-----`,
	},
	"google": {
		`(?i)AIza[0-9A-Za-z\\-_]{35}`,
	},
	"auth": {
		`(?i)Authorization:\s*Bearer\s*[a-zA-Z0-9._-]+`,
		`(?i)api_key`,
		`(?i)apikey`,
		`(?i)secret`,
		`(?i)password`,
	},
}
