// Package asuser runs commands as the agent user of a workspace and asks the root helper
// astra-priv to prepare new workspaces (docs/adr/0005). The daemon never opens a
// workspace's work folder itself: git and every other repository command run as the agent
// user through sudo, with an environment listed here in full.
//
// In local mode (development on a laptop) there are no agent users: commands run as the
// caller under Root, and preparing a workspace only creates its folders.
package asuser
