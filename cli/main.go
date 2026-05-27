package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// ANSI Terminal Formatting Constants
const (
	RED       = "\033[0;31m"
	GREEN     = "\033[0;32m"
	YELLOW    = "\033[0;33m"
	BLUE      = "\033[0;34m"
	CYAN      = "\033[0;36m"
	NC        = "\033[0m" // No Color
	BOLD      = "\033[1m"
	CHECKMARK = "\033[0;32m\xE2\x9C\x94\033[0m"
)

var (
	execDir    string
	sitesFile  string
	composeFile string
)

func init() {
	// Dynamically resolve real execution directory (handles global symlinks)
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("%s[ERROR] Failed to get executable path: %v%s\n", RED, err, NC)
		os.Exit(1)
	}

	realPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		realPath = execPath
	}

	execDir = filepath.Dir(realPath)
	sitesFile = filepath.Join(execDir, "sites.txt")
	composeFile = filepath.Join(execDir, "docker-compose.yml")

	// Ensure sites.txt exists
	if _, err := os.Stat(sitesFile); os.IsNotExist(err) {
		f, _ := os.Create(sitesFile)
		f.Close()
	}
}

func printHeader() {
	fmt.Printf("%s%s⚡ LaraDock Compiled CLI Utility ⚡%s\n", BOLD, CYAN, NC)
}

func showHelp() {
	printHeader()
	fmt.Printf("Usage: %slaradock <command> [arguments]%s\n\n", BOLD, NC)
	fmt.Printf("%sCommands:%s\n", BOLD, NC)
	fmt.Printf("  %s%-28s%s List all active registered projects and PHP versions\n", BOLD+GREEN, "list", NC)
	fmt.Printf("  %s%-28s%s Register a new project (Default PHP: php-82)\n", BOLD+GREEN, "add <domain> [php-version]", NC)
	fmt.Printf("  %s%-28s%s Deregister a project and clean up configurations\n", BOLD+GREEN, "remove <domain>", NC)
	fmt.Printf("  %s%-28s%s Start all LaraDock containers in the background\n", BOLD+GREEN, "up", NC)
	fmt.Printf("  %s%-28s%s Stop and remove all containers\n", BOLD+GREEN, "down", NC)
	fmt.Printf("  %s%-28s%s Restart Nginx or a specific database service\n", BOLD+GREEN, "restart [service]", NC)
	fmt.Printf("  %s%-28s%s Open an interactive terminal inside a container (e.g. php-82, node)\n", BOLD+GREEN, "ssh <service>", NC)
	fmt.Printf("  %s%-28s%s Trust the local Root Certificate Authority (CA)\n", BOLD+GREEN, "trust", NC)
	fmt.Printf("  %s%-28s%s Install this CLI globally as 'laradock' to your system\n", BOLD+GREEN, "install", NC)
	fmt.Printf("  %s%-28s%s Remove the global CLI and clean up host configurations\n", BOLD+GREEN, "uninstall", NC)
	fmt.Printf("\n%sExamples:%s\n", BOLD, NC)
	fmt.Printf("  laradock add blog.test php-84\n")
	fmt.Printf("  laradock ssh php-82\n")
	fmt.Printf("  laradock restart nginx\n\n")
}

func cmdList() {
	printHeader()
	fmt.Printf("%s--------------------------------------------------------------------------------%s\n", CYAN, NC)
	fmt.Printf("%s%-20s %-35s %-15s%s\n", BOLD, "Project Name", "Domain URL", "PHP Version", NC)
	fmt.Printf("%s--------------------------------------------------------------------------------%s\n", CYAN, NC)

	file, err := os.Open(sitesFile)
	if err != nil {
		fmt.Printf("%s[ERROR] Failed to read sites.txt: %v%s\n", RED, err, NC)
		return
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse protocol, domain, and PHP version
		// e.g. https://upms.test(php-82)
		parts := strings.Split(line, "(")
		domain := parts[0]
		domain = strings.Replace(domain, "https://", "", 1)
		domain = strings.Replace(domain, "http://", "", 1)

		project := strings.Split(domain, ".")[0]

		phpVer := "php-82 (default)"
		if len(parts) > 1 {
			phpVer = strings.TrimSuffix(parts[1], ")")
		}

		fmt.Printf("%-20s %-35s %-15s\n", project, "https://"+domain, phpVer)
		count++
	}

	if count == 0 {
		fmt.Printf("%sNo active projects registered. Add one using: laradock add <project>.test%s\n", YELLOW, NC)
	}
	fmt.Printf("%s--------------------------------------------------------------------------------%s\n", CYAN, NC)
}

func cmdAdd(domain, phpVer string) {
	if domain == "" {
		fmt.Printf("%s[ERROR] Missing domain parameter.%s\n", RED, NC)
		fmt.Printf("Usage: laradock add <domain> [php-version]\n")
		os.Exit(1)
	}

	if phpVer == "" {
		phpVer = "php-82"
	}

	// Standardize domain input (strip protocol/whitespace)
	domain = strings.Replace(domain, "https://", "", 1)
	domain = strings.Replace(domain, "http://", "", 1)
	domain = strings.TrimSpace(domain)

	// Validate PHP Version
	matched, _ := regexp.MatchString("^(php-82|php-84|php-85)$", phpVer)
	if !matched {
		fmt.Printf("%s[ERROR] Invalid PHP version: '%s'.%s\n", RED, phpVer, NC)
		fmt.Printf("Supported versions: %sphp-82, php-84, php-85%s\n", BOLD, NC)
		os.Exit(1)
	}

	// Check if already registered
	fileBytes, err := os.ReadFile(sitesFile)
	if err == nil && strings.Contains(string(fileBytes), domain) {
		fmt.Printf("%s[WARNING] Domain '%s' is already registered in sites.txt.%s\n", YELLOW, domain, NC)
		os.Exit(0)
	}

	fmt.Printf("%s[INFO] Registering https://%s (%s)...%s\n", CYAN, domain, phpVer, NC)

	// Append to registry
	f, err := os.OpenFile(sitesFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("%s[ERROR] Failed to write to sites.txt: %v%s\n", RED, err, NC)
		os.Exit(1)
	}
	defer f.Close()

	if _, err := f.WriteString(fmt.Sprintf("https://%s(%s)\n", domain, phpVer)); err != nil {
		fmt.Printf("%s[ERROR] Failed to write to sites.txt: %v%s\n", RED, err, NC)
		os.Exit(1)
	}

	fmt.Printf("%s%s Appended to sites.txt successfully.%s\n", GREEN, CHECKMARK, NC)

	syncEnvironment()
}

func cmdRemove(domain string) {
	if domain == "" {
		fmt.Printf("%s[ERROR] Missing domain parameter.%s\n", RED, NC)
		fmt.Printf("Usage: laradock remove <domain>\n")
		os.Exit(1)
	}

	domain = strings.Replace(domain, "https://", "", 1)
	domain = strings.Replace(domain, "http://", "", 1)
	domain = strings.TrimSpace(domain)

	fileBytes, err := os.ReadFile(sitesFile)
	if err != nil || !strings.Contains(string(fileBytes), domain) {
		fmt.Printf("%s[ERROR] Domain '%s' is not registered in sites.txt.%s\n", RED, domain, NC)
		os.Exit(1)
	}

	fmt.Printf("%s[INFO] Deregistering %s...%s\n", CYAN, domain, NC)

	// Remove target line safely
	file, err := os.Open(sitesFile)
	if err != nil {
		fmt.Printf("%s[ERROR] Failed to open sites.txt: %v%s\n", RED, err, NC)
		os.Exit(1)
	}
	defer file.Close()

	var newLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, domain) {
			newLines = append(newLines, line)
		}
	}

	newContent := strings.Join(newLines, "\n") + "\n"
	if err := os.WriteFile(sitesFile, []byte(newContent), 0644); err != nil {
		fmt.Printf("%s[ERROR] Failed to clean sites.txt: %v%s\n", RED, err, NC)
		os.Exit(1)
	}

	fmt.Printf("%s%s Removed from sites.txt successfully.%s\n", GREEN, CHECKMARK, NC)

	syncEnvironment()
}

func syncEnvironment() {
	fmt.Printf("%s[INFO] Triggering LaraDock Setup to compile SSL, update configs, and sync hosts...%s\n", CYAN, NC)
	
	setupCmd := exec.Command("docker", "compose", "-f", composeFile, "up", "-d", "--build", "setup")
	setupCmd.Stdout = os.Stdout
	setupCmd.Stderr = os.Stderr
	if err := setupCmd.Run(); err != nil {
		fmt.Printf("%s[ERROR] Setup container failed to complete: %v%s\n", RED, err, NC)
		os.Exit(1)
	}

	fmt.Printf("%s[INFO] Restarting Nginx to load changes...%s\n", CYAN, NC)
	nginxCmd := exec.Command("docker", "compose", "-f", composeFile, "restart", "nginx")
	nginxCmd.Stdout = os.Stdout
	nginxCmd.Stderr = os.Stderr
	if err := nginxCmd.Run(); err != nil {
		fmt.Printf("%s[ERROR] Failed to restart Nginx: %v%s\n", RED, err, NC)
		os.Exit(1)
	}

	fmt.Printf("%s%s Environment fully updated and synchronized!%s\n", GREEN, CHECKMARK, NC)
}

func cmdUp() {
	fmt.Printf("%s[INFO] Spin up all LaraDock services...%s\n", CYAN, NC)
	cmd := exec.Command("docker", "compose", "-f", composeFile, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("%s[ERROR] Failed to spin up services: %v%s\n", RED, err, NC)
		os.Exit(1)
	}
	fmt.Printf("%s%s LaraDock is running in background.%s\n", GREEN, CHECKMARK, NC)
}

func cmdDown() {
	fmt.Printf("%s[INFO] Bringing down all LaraDock services...%s\n", CYAN, NC)
	cmd := exec.Command("docker", "compose", "-f", composeFile, "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("%s[ERROR] Failed to bring down services: %v%s\n", RED, err, NC)
		os.Exit(1)
	}
	fmt.Printf("%s%s LaraDock stopped.%s\n", GREEN, CHECKMARK, NC)
}

func cmdRestart(service string) {
	var cmd *exec.Cmd
	if service != "" {
		fmt.Printf("%s[INFO] Restarting service: %s...%s\n", CYAN, service, NC)
		cmd = exec.Command("docker", "compose", "-f", composeFile, "restart", service)
	} else {
		fmt.Printf("%s[INFO] Restarting all services...%s\n", CYAN, NC)
		cmd = exec.Command("docker", "compose", "-f", composeFile, "restart")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("%s[ERROR] Restart failed: %v%s\n", RED, err, NC)
		os.Exit(1)
	}
	fmt.Printf("%s%s Restart completed.%s\n", GREEN, CHECKMARK, NC)
}

func cmdSsh(service string) {
	if service == "" {
		fmt.Printf("%s[ERROR] Missing container service name.%s\n", RED, NC)
		fmt.Printf("Usage: laradock ssh <service>  (e.g., php-82, php-84, node, postgres)\n")
		os.Exit(1)
	}

	fmt.Printf("%s[INFO] Connecting to container: %s...%s\n", CYAN, service, NC)
	
	// Must connect stdin, stdout, and stderr for interactive SSH terminal session
	cmd := exec.Command("docker", "compose", "-f", composeFile, "exec", service, "sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("%s[ERROR] Connection closed: %v%s\n", RED, err, NC)
		os.Exit(1)
	}
}

func cmdTrust() {
	caFile := filepath.Join(execDir, "mkcert-ca", "rootCA.pem")
	if _, err := os.Stat(caFile); os.IsNotExist(err) {
		fmt.Printf("%s[ERROR] Local CA rootCA.pem not found. Make sure the setup container has run at least once.%s\n", RED, NC)
		os.Exit(1)
	}

	fmt.Printf("%s[INFO] Installing LaraDock Certificate Authority to host machine's trusted database...%s\n", CYAN, NC)

	// Check which distribution paths and tools are available on host
	var copyCmd *exec.Cmd
	var updateCmd *exec.Cmd

	if _, err := os.Stat("/etc/ca-certificates/trust-source/anchors"); err == nil {
		// Arch / Manjaro
		exec.Command("sudo", "mkdir", "-p", "/etc/ca-certificates/trust-source/anchors").Run()
		copyCmd = exec.Command("sudo", "cp", caFile, "/etc/ca-certificates/trust-source/anchors/laradock_rootCA.crt")
		
		if _, err := exec.LookPath("trust"); err == nil {
			updateCmd = exec.Command("sudo", "trust", "extract-compat")
		} else if _, err := exec.LookPath("update-ca-certificates"); err == nil {
			updateCmd = exec.Command("sudo", "update-ca-certificates")
		}
	} else if _, err := os.Stat("/etc/pki/ca-trust/source/anchors"); err == nil {
		// Fedora / RHEL
		exec.Command("sudo", "mkdir", "-p", "/etc/pki/ca-trust/source/anchors").Run()
		copyCmd = exec.Command("sudo", "cp", caFile, "/etc/pki/ca-trust/source/anchors/laradock_rootCA.crt")
		
		if _, err := exec.LookPath("update-ca-trust"); err == nil {
			updateCmd = exec.Command("sudo", "update-ca-trust", "extract")
		}
	} else {
		// Debian / Ubuntu / Suse / Alpine fallback
		exec.Command("sudo", "mkdir", "-p", "/usr/local/share/ca-certificates").Run()
		copyCmd = exec.Command("sudo", "cp", caFile, "/usr/local/share/ca-certificates/laradock_rootCA.crt")
		
		if _, err := exec.LookPath("update-ca-certificates"); err == nil {
			updateCmd = exec.Command("sudo", "update-ca-certificates")
		}
	}

	if copyCmd != nil {
		copyCmd.Stdout = os.Stdout
		copyCmd.Stderr = os.Stderr
		copyCmd.Run()
	}

	if updateCmd != nil {
		updateCmd.Stdout = os.Stdout
		updateCmd.Stderr = os.Stderr
		updateCmd.Run()
	} else {
		fmt.Printf("%s[WARNING] No standard update-ca tool found. Mapped CA cert to anchors directory.%s\n", YELLOW, NC)
	}

	fmt.Printf("%s%s Trust established on host system!%s\n", GREEN, CHECKMARK, NC)
	fmt.Printf("%s[NOTE] If you use Firefox, you must still import this file manually inside Firefox certificate settings:%s\n", YELLOW, NC)
	fmt.Printf("%s  %s%s\n", BOLD, caFile, NC)
}

func cmdInstall() {
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("%s[ERROR] Failed to get executable path: %v%s\n", RED, err, NC)
		os.Exit(1)
	}

	fmt.Printf("%s[INFO] Installing 'laradock' globally to /usr/local/bin...%s\n", CYAN, NC)
	exec.Command("sudo", "ln", "-sf", execPath, "/usr/local/bin/laradock").Run()

	// Install UNIX Manual Page
	manSource := filepath.Join(execDir, "laradock.1")
	if _, err := os.Stat(manSource); err == nil {
		fmt.Printf("%s[INFO] Installing 'laradock' manual page...%s\n", CYAN, NC)
		exec.Command("sudo", "mkdir", "-p", "/usr/local/share/man/man1/").Run()
		exec.Command("sudo", "ln", "-sf", manSource, "/usr/local/share/man/man1/laradock.1").Run()
		
		if _, err := exec.LookPath("mandb"); err == nil {
			exec.Command("sudo", "mandb").Run()
		}
	}

	// Shell completion linkages
	// 1. Bash
	bashSource := filepath.Join(execDir, "laradock.bash")
	if _, err := os.Stat(bashSource); err == nil {
		if _, err := os.Stat("/usr/share/bash-completion/completions"); err == nil {
			fmt.Printf("%s[INFO] Installing Bash autocompletion...%s\n", CYAN, NC)
			exec.Command("sudo", "ln", "-sf", bashSource, "/usr/share/bash-completion/completions/laradock").Run()
		} else if _, err := os.Stat("/etc/bash_completion.d"); err == nil {
			fmt.Printf("%s[INFO] Installing Bash autocompletion (legacy path)...%s\n", CYAN, NC)
			exec.Command("sudo", "ln", "-sf", bashSource, "/etc/bash_completion.d/laradock").Run()
		}
	}

	// 2. Zsh
	zshSource := filepath.Join(execDir, "_laradock")
	if _, err := os.Stat(zshSource); err == nil {
		if _, err := os.Stat("/usr/local/share/zsh/site-functions"); err == nil {
			fmt.Printf("%s[INFO] Installing Zsh autocompletion (local)...%s\n", CYAN, NC)
			exec.Command("sudo", "mkdir", "-p", "/usr/local/share/zsh/site-functions/").Run()
			exec.Command("sudo", "ln", "-sf", zshSource, "/usr/local/share/zsh/site-functions/_laradock").Run()
			exec.Command("sudo", "chmod", "644", "/usr/local/share/zsh/site-functions/_laradock").Run()
		} else if _, err := os.Stat("/usr/share/zsh/site-functions"); err == nil {
			fmt.Printf("%s[INFO] Installing Zsh autocompletion (system)...%s\n", CYAN, NC)
			exec.Command("sudo", "ln", "-sf", zshSource, "/usr/share/zsh/site-functions/_laradock").Run()
			exec.Command("sudo", "chmod", "644", "/usr/share/zsh/site-functions/_laradock").Run()
		}
	}

	// 3. Fish
	fishSource := filepath.Join(execDir, "laradock.fish")
	if _, err := os.Stat(fishSource); err == nil {
		if _, err := os.Stat("/usr/share/fish/vendor_completions.d"); err == nil {
			fmt.Printf("%s[INFO] Installing Fish autocompletion...%s\n", CYAN, NC)
			exec.Command("sudo", "ln", "-sf", fishSource, "/usr/share/fish/vendor_completions.d/laradock.fish").Run()
		} else if _, err := os.Stat("/etc/fish/completions"); err == nil {
			fmt.Printf("%s[INFO] Installing Fish autocompletion (etc path)...%s\n", CYAN, NC)
			exec.Command("sudo", "ln", "-sf", fishSource, "/etc/fish/completions/laradock.fish").Run()
		}
	}

	fmt.Printf("%s%s LaraDock CLI installed globally!%s\n", GREEN, CHECKMARK, NC)
	fmt.Printf("You can now run %slaradock <command>%s from ANY directory on your system!\n", BOLD, NC)
	fmt.Printf("%s[NOTE] Please restart your terminal shell (or run 'exec zsh' / 'exec bash' / 'exec fish') to activate autocompletion!%s\n", YELLOW, NC)
}

func cmdUninstall() {
	fmt.Printf("%s[INFO] Uninstalling LaraDock global CLI and system configurations...%s\n", CYAN, NC)

	// 1. Clean up hosts mappings
	hostsFile := "/etc/hosts"
	blockStart := "# BEGIN LARADOCK DOMAINS"
	blockEnd := "# END LARADOCK DOMAINS"

	fileBytes, err := os.ReadFile(hostsFile)
	if err == nil && strings.Contains(string(fileBytes), blockStart) {
		fmt.Printf("%s[INFO] Removing LaraDock domains from host's /etc/hosts file...%s\n", CYAN, NC)
		
		file, err := os.Open(hostsFile)
		if err == nil {
			defer file.Close()
			var newLines []string
			inBlock := false
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, blockStart) {
					inBlock = true
					continue
				}
				if strings.Contains(line, blockEnd) {
					inBlock = false
					continue
				}
				if !inBlock {
					newLines = append(newLines, line)
				}
			}

			newContent := strings.Join(newLines, "\n") + "\n"
			
			// Write safely to hosts
			tempHosts, err := os.CreateTemp("", "hosts")
			if err == nil {
				tempHosts.Write([]byte(newContent))
				tempHosts.Close()
				exec.Command("sudo", "cp", tempHosts.Name(), hostsFile).Run()
				os.Remove(tempHosts.Name())
			}
		}
	}

	// 2. Clean up CA trust certs
	fmt.Printf("%s[INFO] Removing trusted Certificate Authority...%s\n", CYAN, NC)
	exec.Command("sudo", "rm", "-f", "/usr/local/share/ca-certificates/laradock_rootCA.crt").Run()
	exec.Command("sudo", "rm", "-f", "/etc/pki/ca-trust/source/anchors/laradock_rootCA.crt").Run()
	exec.Command("sudo", "rm", "-f", "/etc/ca-certificates/trust-source/anchors/laradock_rootCA.crt").Run()

	if _, err := exec.LookPath("trust"); err == nil {
		exec.Command("sudo", "trust", "extract-compat").Run()
	} else if _, err := exec.LookPath("update-ca-certificates"); err == nil {
		exec.Command("sudo", "update-ca-certificates").Run()
	} else if _, err := exec.LookPath("update-ca-trust"); err == nil {
		exec.Command("sudo", "update-ca-trust", "extract").Run()
	}

	// 3. Clean up symlinks, man page, and completions
	fmt.Printf("%s[INFO] Removing global symlinks, completions, and manual pages...%s\n", CYAN, NC)
	exec.Command("sudo", "rm", "-f", "/usr/local/bin/laradock").Run()
	exec.Command("sudo", "rm", "-f", "/usr/local/share/man/man1/laradock.1").Run()
	exec.Command("sudo", "rm", "-f", "/usr/share/bash-completion/completions/laradock").Run()
	exec.Command("sudo", "rm", "-f", "/etc/bash_completion.d/laradock").Run()
	exec.Command("sudo", "rm", "-f", "/usr/local/share/zsh/site-functions/_laradock").Run()
	exec.Command("sudo", "rm", "-f", "/usr/share/zsh/site-functions/_laradock").Run()
	exec.Command("sudo", "rm", "-f", "/usr/share/fish/vendor_completions.d/laradock.fish").Run()
	exec.Command("sudo", "rm", "-f", "/etc/fish/completions/laradock.fish").Run()

	if _, err := exec.LookPath("mandb"); err == nil {
		exec.Command("sudo", "mandb").Run()
	}

	fmt.Printf("%s%s LaraDock CLI successfully uninstalled from host!%s\n", GREEN, CHECKMARK, NC)
}

func main() {
	args := os.Args
	if len(args) < 2 {
		showHelp()
		return
	}

	command := args[1]
	switch command {
	case "list":
		cmdList()
	case "add":
		domain := ""
		phpVer := ""
		if len(args) > 2 {
			domain = args[2]
		}
		if len(args) > 3 {
			phpVer = args[3]
		}
		cmdAdd(domain, phpVer)
	case "remove":
		domain := ""
		if len(args) > 2 {
			domain = args[2]
		}
		cmdRemove(domain)
	case "up":
		cmdUp()
	case "down":
		cmdDown()
	case "restart":
		service := ""
		if len(args) > 2 {
			service = args[2]
		}
		cmdRestart(service)
	case "ssh":
		service := ""
		if len(args) > 2 {
			service = args[2]
		}
		cmdSsh(service)
	case "trust":
		cmdTrust()
	case "install":
		cmdInstall()
	case "uninstall":
		cmdUninstall()
	default:
		showHelp()
	}
}
