package cmd

import "github.com/spf13/cobra"

var installCmd = &cobra.Command{
	Use:     "install <pkgs...>",
	Aliases: []string{"i"},
	Short:   "Install packages",
	Args:    cobra.MinimumNArgs(1),
	Example: "  tvpkg install git\n  tvpkg i python",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, mgr, err := resolve()
		if err != nil {
			return err
		}
		return mgr.Install(args...)
	},
}

var removeCmd = &cobra.Command{
	Use:     "remove <pkgs...>",
	Aliases: []string{"r", "rm", "uninstall", "del"},
	Short:   "Remove packages",
	Args:    cobra.MinimumNArgs(1),
	Example: "  tvpkg remove nodejs\n  tvpkg rm python",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, mgr, err := resolve()
		if err != nil {
			return err
		}
		return mgr.Remove(args...)
	},
}

var searchCmd = &cobra.Command{
	Use:     "search <query>",
	Aliases: []string{"s"},
	Short:   "Search packages by name or description",
	Args:    cobra.MinimumNArgs(1),
	Example: "  tvpkg search scanner\n  tvpkg s nmap",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, mgr, err := resolve()
		if err != nil {
			return err
		}
		return mgr.Search(args[0])
	},
}

var updateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"u"},
	Short:   "Update package databases",
	Args:    cobra.NoArgs,
	Example: "  tvpkg update\n  tvpkg u",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, mgr, err := resolve()
		if err != nil {
			return err
		}
		return mgr.Update()
	},
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "List packages",
	Args:    cobra.NoArgs,
	Example: "  tvpkg list\n  tvpkg l --installed",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, mgr, err := resolve()
		if err != nil {
			return err
		}
		if installed, _ := cmd.Flags().GetBool("installed"); installed {
			return mgr.ListInstalled()
		}
		return mgr.List()
	},
}

var infoCmd = &cobra.Command{
	Use:     "info <pkg>",
	Aliases: []string{"show", "sh"},
	Short:   "Show package metadata",
	Args:    cobra.MinimumNArgs(1),
	Example: "  tvpkg info git",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, mgr, err := resolve()
		if err != nil {
			return err
		}
		return mgr.Info(args[0])
	},
}

func init() {
	listCmd.Flags().BoolP("installed", "i", false, "list only installed packages")
}
