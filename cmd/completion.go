package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

	$ source <(drone-email completion bash)

	# To load completions for each session, execute once:
	# Linux:
	$ drone-email completion bash > /etc/bash_completion.d/drone-email
	# macOS:
	$ drone-email completion bash > $(brew --prefix)/etc/bash_completion.d/drone-email

Zsh:

	# If shell completion is not already enabled in your environment,
	# you will need to enable it.  You can execute the following once:

	$ echo "autoload -U compinit; compinit" >> ~/.zshrc

	# To load completions for each session, execute once:
	$ drone-email completion zsh > "${fpath[1]}/_drone-email"

	# You will need to start a new shell for this setup to take effect.

fish:

	$ drone-email completion fish | source

	# To load completions for each session, execute once:
	$ drone-email completion fish > ~/.config/fish/completions/drone-email.fish

PowerShell:

	PS> drone-email completion powershell | Out-String | Invoke-Expression

	# To load completions for every new session, run:
	PS> drone-email completion powershell > drone-email.ps1
	# and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Hidden:                true,
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			_ = cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			_ = cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			_ = cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			_ = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}
