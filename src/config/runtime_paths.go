package config

import "fmt"

type RuntimePathOptions struct {
	GoOS         string
	HomeDir      string
	IsPrivileged bool
	IsContainer  bool
}

type RuntimePaths struct {
	BinaryPath  string
	ConfigDir   string
	ConfigFile  string
	DataDir     string
	CacheDir    string
	LogDir      string
	LogFile     string
	BackupDir   string
	PIDFile     string
	SSLDir      string
	SecurityDir string
	SQLiteDir   string
}

func ResolveRuntimePaths(options RuntimePathOptions) (RuntimePaths, error) {
	if options.IsContainer {
		return RuntimePaths{
			BinaryPath:  "/usr/local/bin/caspbx",
			ConfigDir:   "/config/caspbx/",
			ConfigFile:  "/config/caspbx/server.yml",
			DataDir:     "/data/caspbx/",
			CacheDir:    "/data/caspbx/cache/",
			LogDir:      "/data/log/caspbx/",
			LogFile:     "/data/log/caspbx/server.log",
			BackupDir:   "/data/backups/caspbx/",
			PIDFile:     "/data/caspbx/caspbx.pid",
			SSLDir:      "/config/caspbx/ssl/",
			SecurityDir: "/data/caspbx/security/",
			SQLiteDir:   "/data/db/sqlite/",
		}, nil
	}

	switch options.GoOS {
	case "linux":
		return resolveLinuxRuntimePaths(options)
	case "darwin":
		return resolveDarwinRuntimePaths(options)
	case "freebsd", "openbsd", "netbsd", "dragonfly":
		return resolveBSDRuntimePaths(options)
	case "windows":
		return resolveWindowsRuntimePaths(options)
	default:
		return RuntimePaths{}, fmt.Errorf("unsupported operating system %q", options.GoOS)
	}
}

func resolveLinuxRuntimePaths(options RuntimePathOptions) (RuntimePaths, error) {
	if options.IsPrivileged {
		return RuntimePaths{
			BinaryPath:  "/usr/local/bin/caspbx",
			ConfigDir:   "/etc/casapps/caspbx/",
			ConfigFile:  "/etc/casapps/caspbx/server.yml",
			DataDir:     "/var/lib/casapps/caspbx/",
			CacheDir:    "/var/cache/casapps/caspbx/",
			LogDir:      "/var/log/casapps/caspbx/",
			LogFile:     "/var/log/casapps/caspbx/server.log",
			BackupDir:   "/mnt/Backups/casapps/caspbx/",
			PIDFile:     "/var/run/casapps/caspbx.pid",
			SSLDir:      "/etc/casapps/caspbx/ssl/",
			SecurityDir: "/var/lib/casapps/caspbx/security/",
			SQLiteDir:   "/var/lib/casapps/caspbx/db/",
		}, nil
	}

	if options.HomeDir == "" {
		return RuntimePaths{}, fmt.Errorf("home directory required for non-privileged linux paths")
	}

	return RuntimePaths{
		BinaryPath:  options.HomeDir + "/.local/bin/caspbx",
		ConfigDir:   options.HomeDir + "/.config/casapps/caspbx/",
		ConfigFile:  options.HomeDir + "/.config/casapps/caspbx/server.yml",
		DataDir:     options.HomeDir + "/.local/share/casapps/caspbx/",
		CacheDir:    options.HomeDir + "/.cache/casapps/caspbx/",
		LogDir:      options.HomeDir + "/.local/log/casapps/caspbx/",
		LogFile:     options.HomeDir + "/.local/log/casapps/caspbx/server.log",
		BackupDir:   options.HomeDir + "/.local/share/Backups/casapps/caspbx/",
		PIDFile:     options.HomeDir + "/.local/share/casapps/caspbx/caspbx.pid",
		SSLDir:      options.HomeDir + "/.config/casapps/caspbx/ssl/",
		SecurityDir: options.HomeDir + "/.local/share/casapps/caspbx/security/",
		SQLiteDir:   options.HomeDir + "/.local/share/casapps/caspbx/db/",
	}, nil
}

func resolveDarwinRuntimePaths(options RuntimePathOptions) (RuntimePaths, error) {
	if options.IsPrivileged {
		return RuntimePaths{
			BinaryPath:  "/usr/local/bin/caspbx",
			ConfigDir:   "/Library/Application Support/casapps/caspbx/",
			ConfigFile:  "/Library/Application Support/casapps/caspbx/server.yml",
			DataDir:     "/Library/Application Support/casapps/caspbx/data/",
			CacheDir:    "/Library/Caches/casapps/caspbx/",
			LogDir:      "/Library/Logs/casapps/caspbx/",
			LogFile:     "/Library/Logs/casapps/caspbx/server.log",
			BackupDir:   "/Library/Backups/casapps/caspbx/",
			PIDFile:     "/var/run/casapps/caspbx.pid",
			SSLDir:      "/Library/Application Support/casapps/caspbx/ssl/",
			SecurityDir: "/Library/Application Support/casapps/caspbx/data/security/",
			SQLiteDir:   "/Library/Application Support/casapps/caspbx/db/",
		}, nil
	}

	if options.HomeDir == "" {
		return RuntimePaths{}, fmt.Errorf("home directory required for non-privileged darwin paths")
	}

	return RuntimePaths{
		BinaryPath:  options.HomeDir + "/bin/caspbx",
		ConfigDir:   options.HomeDir + "/Library/Application Support/casapps/caspbx/",
		ConfigFile:  options.HomeDir + "/Library/Application Support/casapps/caspbx/server.yml",
		DataDir:     options.HomeDir + "/Library/Application Support/casapps/caspbx/",
		CacheDir:    options.HomeDir + "/Library/Caches/casapps/caspbx/",
		LogDir:      options.HomeDir + "/Library/Logs/casapps/caspbx/",
		LogFile:     options.HomeDir + "/Library/Logs/casapps/caspbx/server.log",
		BackupDir:   options.HomeDir + "/Library/Backups/casapps/caspbx/",
		PIDFile:     options.HomeDir + "/Library/Application Support/casapps/caspbx/caspbx.pid",
		SSLDir:      options.HomeDir + "/Library/Application Support/casapps/caspbx/ssl/",
		SecurityDir: options.HomeDir + "/Library/Application Support/casapps/caspbx/data/security/",
		SQLiteDir:   options.HomeDir + "/Library/Application Support/casapps/caspbx/db/",
	}, nil
}

func resolveBSDRuntimePaths(options RuntimePathOptions) (RuntimePaths, error) {
	if options.IsPrivileged {
		return RuntimePaths{
			BinaryPath:  "/usr/local/bin/caspbx",
			ConfigDir:   "/usr/local/etc/casapps/caspbx/",
			ConfigFile:  "/usr/local/etc/casapps/caspbx/server.yml",
			DataDir:     "/var/db/casapps/caspbx/",
			CacheDir:    "/var/cache/casapps/caspbx/",
			LogDir:      "/var/log/casapps/caspbx/",
			LogFile:     "/var/log/casapps/caspbx/server.log",
			BackupDir:   "/var/backups/casapps/caspbx/",
			PIDFile:     "/var/run/casapps/caspbx.pid",
			SSLDir:      "/usr/local/etc/casapps/caspbx/ssl/",
			SecurityDir: "/var/db/casapps/caspbx/security/",
			SQLiteDir:   "/var/db/casapps/caspbx/db/",
		}, nil
	}

	if options.HomeDir == "" {
		return RuntimePaths{}, fmt.Errorf("home directory required for non-privileged bsd paths")
	}

	return RuntimePaths{
		BinaryPath:  options.HomeDir + "/.local/bin/caspbx",
		ConfigDir:   options.HomeDir + "/.config/casapps/caspbx/",
		ConfigFile:  options.HomeDir + "/.config/casapps/caspbx/server.yml",
		DataDir:     options.HomeDir + "/.local/share/casapps/caspbx/",
		CacheDir:    options.HomeDir + "/.cache/casapps/caspbx/",
		LogDir:      options.HomeDir + "/.local/log/casapps/caspbx/",
		LogFile:     options.HomeDir + "/.local/log/casapps/caspbx/server.log",
		BackupDir:   options.HomeDir + "/.local/share/Backups/casapps/caspbx/",
		PIDFile:     options.HomeDir + "/.local/share/casapps/caspbx/caspbx.pid",
		SSLDir:      options.HomeDir + "/.config/casapps/caspbx/ssl/",
		SecurityDir: options.HomeDir + "/.local/share/casapps/caspbx/security/",
		SQLiteDir:   options.HomeDir + "/.local/share/casapps/caspbx/db/",
	}, nil
}

func resolveWindowsRuntimePaths(options RuntimePathOptions) (RuntimePaths, error) {
	if options.IsPrivileged {
		return RuntimePaths{
			BinaryPath:  `C:\Program Files\casapps\caspbx\caspbx.exe`,
			ConfigDir:   `%ProgramData%\casapps\caspbx\`,
			ConfigFile:  `%ProgramData%\casapps\caspbx\server.yml`,
			DataDir:     `%ProgramData%\casapps\caspbx\data\`,
			CacheDir:    `%ProgramData%\casapps\caspbx\cache\`,
			LogDir:      `%ProgramData%\casapps\caspbx\logs\`,
			LogFile:     `%ProgramData%\casapps\caspbx\logs\server.log`,
			BackupDir:   `%ProgramData%\Backups\casapps\caspbx\`,
			PIDFile:     `%ProgramData%\casapps\caspbx\caspbx.pid`,
			SSLDir:      `%ProgramData%\casapps\caspbx\ssl\`,
			SecurityDir: `%ProgramData%\casapps\caspbx\data\security\`,
			SQLiteDir:   `%ProgramData%\casapps\caspbx\db\`,
		}, nil
	}

	if options.HomeDir == "" {
		return RuntimePaths{}, fmt.Errorf("home directory required for non-privileged windows paths")
	}

	return RuntimePaths{
		BinaryPath:  options.HomeDir + `\AppData\Local\casapps\caspbx\caspbx.exe`,
		ConfigDir:   options.HomeDir + `\AppData\Roaming\casapps\caspbx\`,
		ConfigFile:  options.HomeDir + `\AppData\Roaming\casapps\caspbx\server.yml`,
		DataDir:     options.HomeDir + `\AppData\Local\casapps\caspbx\`,
		CacheDir:    options.HomeDir + `\AppData\Local\casapps\caspbx\cache\`,
		LogDir:      options.HomeDir + `\AppData\Local\casapps\caspbx\logs\`,
		LogFile:     options.HomeDir + `\AppData\Local\casapps\caspbx\logs\server.log`,
		BackupDir:   options.HomeDir + `\AppData\Local\Backups\casapps\caspbx\`,
		PIDFile:     options.HomeDir + `\AppData\Local\casapps\caspbx\caspbx.pid`,
		SSLDir:      options.HomeDir + `\AppData\Roaming\casapps\caspbx\ssl\`,
		SecurityDir: options.HomeDir + `\AppData\Local\casapps\caspbx\security\`,
		SQLiteDir:   options.HomeDir + `\AppData\Local\casapps\caspbx\db\`,
	}, nil
}
