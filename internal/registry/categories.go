package registry

// categories contains all registered cache categories
var categories = []CacheCategory{
	// ==================== macOS System ====================
	{
		ID:          "macos-user-cache",
		Name:        "macOS User Cache",
		Description: "User-level application caches",
		Platforms:   []string{"macos"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/Library/Caches"},
		},
	},
	{
		ID:          "macos-system-cache",
		Name:        "macOS System Cache",
		Description: "System-level caches (use with caution)",
		Platforms:   []string{"macos"},
		SafetyLevel: Caution,
		Paths: []PathRule{
			{Path: "/Library/Caches"},
		},
	},
	{
		ID:          "macos-logs",
		Name:        "macOS Logs",
		Description: "System and application logs",
		Platforms:   []string{"macos"},
		SafetyLevel: Caution,
		Paths: []PathRule{
			{Path: "~/Library/Logs"},
			{Path: "/Library/Logs"},
		},
	},

	// ==================== Xcode (macOS) ====================
	{
		ID:          "xcode-deriveddata",
		Name:        "Xcode DerivedData",
		Description: "Build intermediates and compiled objects",
		Platforms:   []string{"macos"},
		SafetyLevel: Caution,
		Paths: []PathRule{
			{Path: "~/Library/Developer/Xcode/DerivedData"},
		},
		CleanCmd: "rm -rf ~/Library/Developer/Xcode/DerivedData/*",
	},
	{
		ID:          "xcode-simulator",
		Name:        "Xcode Simulator",
		Description: "iOS Simulator data and apps",
		Platforms:   []string{"macos"},
		SafetyLevel: Caution,
		Paths: []PathRule{
			{Path: "~/Library/Developer/CoreSimulator"},
		},
	},
	{
		ID:          "xcode-archives",
		Name:        "Xcode Archives",
		Description: "Archived builds",
		Platforms:   []string{"macos"},
		SafetyLevel: Caution,
		Paths: []PathRule{
			{Path: "~/Library/Developer/Xcode/Archives"},
		},
	},
	{
		ID:          "xcode-devicesupport",
		Name:        "Xcode DeviceSupport",
		Description: "iOS device support files",
		Platforms:   []string{"macos"},
		SafetyLevel: Caution,
		Paths: []PathRule{
			{Path: "~/Library/Developer/Xcode/iOS DeviceSupport"},
		},
	},

	// ==================== Docker ====================
	{
		ID:          "docker",
		Name:        "Docker",
		Description: "Docker images, containers, and build cache",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Caution,
		Paths: []PathRule{
			{Path: "/var/lib/docker/containers"},
			{Path: "/var/lib/docker/overlay2"},
		},
		CleanCmd: "docker system prune -a",
	},

	// ==================== Python ====================
	{
		ID:          "python-pip-cache",
		Name:        "Python pip Cache",
		Description: "pip downloaded packages",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/Library/Caches/pip"},
			{Path: "~/.cache/pip"},
		},
		CleanCmd: "pip cache purge",
	},
	{
		ID:          "python-pyenv",
		Name:        "Python pyenv",
		Description: "Multiple Python versions managed by pyenv",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Caution,
		Paths: []PathRule{
			{Path: "~/.pyenv/versions"},
		},
	},
	{
		ID:          "python-venv",
		Name:        "Python Virtual Environments",
		Description: "Project virtual environments (requires project path)",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Caution,
		Paths: []PathRule{
			{Path: "~/.*/venv"},
			{Path: "~/.virtualenvs"},
		},
	},

	// ==================== Node.js ====================
	{
		ID:          "npm-cache",
		Name:        "npm Cache",
		Description: "npm downloaded packages",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/.npm"},
		},
		CleanCmd: "npm cache clean --force",
	},
	{
		ID:          "yarn-cache",
		Name:        "Yarn Cache",
		Description: "Yarn downloaded packages",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/.cache/yarn"},
			{Path: "~/.yarn"},
		},
		CleanCmd: "yarn cache clean",
	},
	{
		ID:          "node-modules",
		Name:        "node_modules",
		Description: "Project node_modules (requires project path)",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Caution,
		Paths: []PathRule{
			{Path: "~/node_modules"},
		},
	},

	// ==================== Go ====================
	{
		ID:          "go-modules",
		Name:        "Go Module Cache",
		Description: "Go downloaded modules",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/go/pkg/mod"},
			{Path: "$GOPATH/pkg/mod"},
		},
		CleanCmd: "go clean -modcache",
	},
	{
		ID:          "go-build-cache",
		Name:        "Go Build Cache",
		Description: "Go build cache",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/.cache/go-build"},
		},
		CleanCmd: "go clean -cache",
	},

	// ==================== Rust ====================
	{
		ID:          "cargo-registry",
		Name:        "Cargo Registry",
		Description: "Rust crate registry",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/.cargo/registry"},
		},
		CleanCmd: "cargo cache --autoclean",
	},
	{
		ID:          "cargo-git",
		Name:        "Cargo Git Cache",
		Description: "Cargo git checkouts",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/.cargo/git"},
		},
	},

	// ==================== Java ====================
	{
		ID:          "maven-repository",
		Name:        "Maven Local Repository",
		Description: "Maven downloaded dependencies",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/.m2/repository"},
		},
		CleanCmd: "mvn dependency:purge-local-repository",
	},
	{
		ID:          "gradle-cache",
		Name:        "Gradle Cache",
		Description: "Gradle caches and dependencies",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/.gradle/caches"},
		},
	},

	// ==================== Ruby ====================
	{
		ID:          "gem-cache",
		Name:        "RubyGems Cache",
		Description: "Ruby gem cache",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/.gem"},
			{Path: "~/.cache/bundle"},
		},
		CleanCmd: "gem cleanup",
	},

	// ==================== Elixir ====================
	{
		ID:          "hex-packages",
		Name:        "Hex Packages",
		Description: "Elixir/Hex package cache",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/.hex"},
		},
	},

	// ==================== Homebrew ====================
	{
		ID:          "homebrew-cache",
		Name:        "Homebrew Cache",
		Description: "Homebrew downloaded bottles",
		Platforms:   []string{"macos"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/Library/Caches/Homebrew"},
			{Path: "/Library/Caches/Homebrew"},
		},
		CleanCmd: "brew cleanup",
	},

	// ==================== Linux Package Managers ====================
	{
		ID:          "apt-cache",
		Name:        "APT Cache",
		Description: "Debian/Ubuntu package cache",
		Platforms:   []string{"linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "/var/cache/apt/archives"},
		},
		CleanCmd: "apt clean",
	},
	{
		ID:          "dnf-cache",
		Name:        "DNF Cache",
		Description: "Fedora/RHEL package cache",
		Platforms:   []string{"linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "/var/cache/dnf"},
		},
		CleanCmd: "dnf clean all",
	},
	{
		ID:          "pacman-cache",
		Name:        "Pacman Cache",
		Description: "Arch Linux package cache",
		Platforms:   []string{"linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "/var/cache/pacman/pkg"},
		},
		CleanCmd: "pacman -Scc",
	},

	// ==================== Browsers ====================
	{
		ID:          "chrome-cache",
		Name:        "Chrome Cache",
		Description: "Google Chrome browser cache",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/Library/Caches/Google/Chrome"},
			{Path: "~/.cache/google-chrome"},
		},
	},
	{
		ID:          "firefox-cache",
		Name:        "Firefox Cache",
		Description: "Mozilla Firefox browser cache",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/Library/Caches/Firefox"},
			{Path: "~/.cache/mozilla/firefox"},
		},
	},
	{
		ID:          "safari-cache",
		Name:        "Safari Cache",
		Description: "Safari browser cache",
		Platforms:   []string{"macos"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/Library/Caches/com.apple.Safari"},
		},
	},

	// ==================== Desktop Environment ====================
	{
		ID:          "thumbnails",
		Name:        "Thumbnails",
		Description: "Image and video thumbnails",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/Library/Caches/com.apple.finder"},
			{Path: "~/.cache/thumbnails"},
		},
	},
	{
		ID:          "fontconfig-cache",
		Name:        "Font Cache",
		Description: "Fontconfig font cache",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/.cache/fontconfig"},
			{Path: "/var/cache/fontconfig"},
		},
	},

	// ==================== IDEs ====================
	{
		ID:          "vscode-cache",
		Name:        "VSCode Cache",
		Description: "Visual Studio Code cache",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/Library/Caches/com.microsoft.VSCode"},
			{Path: "~/.config/Code/Cache"},
			{Path: "~/.cache/Code"},
		},
	},
	{
		ID:          "jetbrains-cache",
		Name:        "JetBrains Cache",
		Description: "JetBrains IDEs cache (IntelliJ, PyCharm, etc.)",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/Library/Caches/JetBrains"},
			{Path: "~/.cache/JetBrains"},
		},
	},

	// ==================== GitHub CLI ====================
	{
		ID:          "gh-cache",
		Name:        "GitHub CLI Cache",
		Description: "GitHub CLI cached data",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/.cache/gh"},
		},
	},

	// ==================== Terraform ====================
	{
		ID:          "terraform-cache",
		Name:        "Terraform Cache",
		Description: "Terraform provider cache",
		Platforms:   []string{"macos", "linux"},
		SafetyLevel: Safe,
		Paths: []PathRule{
			{Path: "~/.terraform.d/plugins"},
			{Path: "~/.cache/terraform"},
		},
	},
}
