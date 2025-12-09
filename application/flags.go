package application

import "github.com/spf13/cobra"

type PersistentFlags struct {
	AddJsonFlag    bool
	AddQuietFlag   bool
	AddNoColorFlag bool
	AddVerboseFlag bool
	AddVersionFlag bool
}

func (f PersistentFlags) configureFlags(cmd *cobra.Command) {
	f.configureVersionFlag(cmd)                           // Configure app for version information
	f.configureOutputFlags(cmd)                           // Configure verbosity
	f.configureLoggingFlags(cmd)                          // Configure logging
	f.configureExclusions(cmd)                            // Configure mutually exclusive flags
	cmd.PersistentFlags().SetNormalizeFunc(normalizeFunc) // normalize persistent flags
}

func (f PersistentFlags) configureExclusions(cmd *cobra.Command) {
	if f.AddNoColorFlag {
		cmd.MarkFlagsMutuallyExclusive(noColorFlag.Name(), logTypeFlag.Name())
	}

	if f.AddJsonFlag && f.AddNoColorFlag {
		cmd.MarkFlagsMutuallyExclusive(jsonOutputFlag.Name(), noColorFlag.Name())
	}

	if f.AddQuietFlag && f.AddNoColorFlag {
		cmd.MarkFlagsMutuallyExclusive(quietFlag.Name(), noColorFlag.Name())
	}

	if f.AddQuietFlag && f.AddVerboseFlag {
		cmd.MarkFlagsMutuallyExclusive(quietFlag.Name(), verboseFlag.Name())
	}

	if f.AddQuietFlag && f.AddJsonFlag {
		cmd.MarkFlagsMutuallyExclusive(jsonOutputFlag.Name(), quietFlag.Name())
	}
}

func (f PersistentFlags) configureLoggingFlags(cmd *cobra.Command) {
	addLogLevelFlag(cmd)
	addLogOutputFlag(cmd)
	addLogTypeFlag(cmd)
}

func (f PersistentFlags) configureOutputFlags(cmd *cobra.Command) {
	if f.AddJsonFlag {
		addJsonOutputFlag(cmd)
	}

	if f.AddNoColorFlag {
		addNoColorFlag(cmd)
	}

	if f.AddVerboseFlag {
		addVerboseFlag(cmd)
	}

	if f.AddQuietFlag {
		addQuietFlag(cmd)
	}
}

func (f PersistentFlags) configureVersionFlag(cmd *cobra.Command) {
	if f.AddVersionFlag {
		addVersionFlag(cmd)
	}
}
