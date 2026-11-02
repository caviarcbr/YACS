package tui

type menuItem struct {
	title       string
	description string
}

var menuItems = []menuItem{
	{
		title:       "Install a server",
		description: "Set up a new ownCloud Infinite Scale deployment.",
	},
	{
		title:       "Restore an installation",
		description: "Bring back an existing YACS deployment.",
	},
	{
		title:       "Uninstall",
		description: "Remove your YACS deployment.",
	},
}
