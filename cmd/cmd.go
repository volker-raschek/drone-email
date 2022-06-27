package cmd

import (
	"fmt"
	"os"
	"strings"

	"git.cryptic.systems/volker.raschek/drone-email-docker/pkg/domain"
	"git.cryptic.systems/volker.raschek/drone-email-docker/pkg/flags"
	"git.cryptic.systems/volker.raschek/drone-email-docker/pkg/mail"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	// The name of our config file, without the file extension because viper
	// supports many different config file languages.
	defaultConfigFilename  = "config"
	defaultConfigExtension = "yaml"

	// The environment variable prefix of all environment variables bound to our command line flags.
	// For example, --number is bound to STING_NUMBER.
	envPrefix = ""
)

func Execute(version string) error {
	rootCmd := &cobra.Command{
		Use: "drone-email",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initializeConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			vars, err := newHTMLTemplateVarsByCommand(cmd)
			if err != nil {
				return fmt.Errorf("failed to initialize new html template vars: %w", err)
			}

			smtpSettings, err := newSMTPSettingsByCommand(cmd)
			if err != nil {
				return fmt.Errorf("failed to initialize new config vars: %w", err)
			}

			recipients, err := cmd.Flags().GetStringArray(flags.SMTP_TO_ADDRESSES)
			if err != nil {
				return fmt.Errorf("failed to detect value of %s: %w", flags.SMTP_TO_ADDRESSES, err)
			}

			err = mail.NewPlugin(smtpSettings).Exec(cmd.Context(), recipients, vars)
			if err != nil {
				return fmt.Errorf("failed to execute mail plugin: %w", err)
			}

			_, err = fmt.Fprint(os.Stdout, "E-Mails successfully sent")
			if err != nil {
				return fmt.Errorf("failed to write message on stdout: %w", err)
			}

			return nil
		},
		SilenceUsage: true,
		Version:      version,
	}

	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to detect hostname: %w", err)
	}

	// Drone environment variables/flags
	// Flags which receive their values from environment variables of the drone
	// CI/CD.
	//
	// The names of the FLags must match the environment variables, otherwise the
	// environment variables will not be bound correctly to the flags.
	rootCmd.Flags().Int64(flags.DRONE_BUILD_CREATED, 0, "Build created")
	rootCmd.Flags().Int64(flags.DRONE_BUILD_FINISHED, 0, "Build finished")
	rootCmd.Flags().Int64(flags.DRONE_BUILD_STARTED, 0, "Build stared")
	rootCmd.Flags().String(flags.DRONE_BUILD_EVENT, "push", "Build event")
	rootCmd.Flags().String(flags.DRONE_BUILD_LINK, "", "Build link")
	rootCmd.Flags().Int(flags.DRONE_BUILD_NUMBER, 0, "Build number")
	rootCmd.Flags().String(flags.DRONE_BUILD_STATUS, "success", "Build status")

	rootCmd.Flags().String(flags.DRONE_COMMIT_SHA, "", "SHA sum of the commit")
	rootCmd.Flags().String(flags.DRONE_COMMIT_REF, "refs/heads/master", "Commit reference")
	rootCmd.Flags().String(flags.DRONE_COMMIT_BRANCH, "master", "Commit branch")
	rootCmd.Flags().String(flags.DRONE_COMMIT_LINK, "", "Link to the commit")
	rootCmd.Flags().String(flags.DRONE_COMMIT_MESSAGE, "", "Commit message")
	rootCmd.Flags().String(flags.DRONE_COMMIT_AUTHOR_NAME, "", "Name of the commit author")
	rootCmd.Flags().String(flags.DRONE_COMMIT_AUTHOR_EMAIL, "", "E-Mail of the commit author")
	rootCmd.Flags().String(flags.DRONE_COMMIT_AUTHOR_AVATAR, "", "Avatar of the commit author")

	rootCmd.Flags().String(flags.DRONE_DEPLOY_TO, "", "Deploy target")

	rootCmd.Flags().String(flags.DRONE_JOB_NUMBER, "", "Job number")
	rootCmd.Flags().String(flags.DRONE_JOB_STATUS, "", "Job status")
	rootCmd.Flags().Int(flags.DRONE_JOB_EXIT_CODE, 0, "Job exit code")
	rootCmd.Flags().Int(flags.DRONE_JOB_STARTED, 0, "Job started")
	rootCmd.Flags().Int(flags.DRONE_JOB_FINISHED, 0, "Job finished")

	rootCmd.Flags().String(flags.DRONE_PREV_BUILD_STATUS, "", "Previous build status")
	rootCmd.Flags().Int(flags.DRONE_PREV_BUILD_NUMBER, 0, "Previous build number")
	rootCmd.Flags().String(flags.DRONE_PREV_COMMIT_SHA, "", "Previous commit sha sum")

	rootCmd.Flags().Int(flags.DRONE_PULL_REQUEST, 0, "Number of pull-request")

	rootCmd.Flags().String(flags.DRONE_REMOTE_URL, "", "Clone URL of the repository")

	rootCmd.Flags().Bool(flags.DRONE_REPO_PRIVATE, true, "Repository is private")
	rootCmd.Flags().Bool(flags.DRONE_REPO_TRUSTED, false, "Repository is trusted")
	rootCmd.Flags().String(flags.DRONE_REPO_AVATAR, "", "Avatar URL of the repository")
	rootCmd.Flags().String(flags.DRONE_REPO_BRANCH, "master", "Branch of the repository")
	rootCmd.Flags().String(flags.DRONE_REPO_LINK, "", "URL to the repository")
	rootCmd.Flags().String(flags.DRONE_REPO_NAME, "", "Name of the repository")
	rootCmd.Flags().String(flags.DRONE_REPO_OWNER, "", "Name of the repository owner")
	rootCmd.Flags().String(flags.DRONE_REPO_SCM, "git", "Source code management provider")
	rootCmd.Flags().String(flags.DRONE_REPO, "", "Full name of the repository")

	rootCmd.Flags().String(flags.DRONE_TAG, "", "Tag")

	rootCmd.Flags().Bool(flags.DRONE_YAML_SIGNED, false, "YAML is signed")
	rootCmd.Flags().Bool(flags.DRONE_YAML_VERIFIED, false, "YAML is verified")

	// MAIL SETTINGS
	rootCmd.Flags().Bool(flags.SMTP_START_TLS, mail.DefaultSMTPStartTLS, "Use StartTLS instead of SSL")
	rootCmd.Flags().Bool(flags.SMTP_TLS_INSECURE_SKIP_VERIFY, mail.DefaultSMTPTLSInsecureSkipVerify, "Trust insecure TLS certificates")
	rootCmd.Flags().Int(flags.SMTP_PORT, mail.DefaultSMTPPort, "SMTP-Port")
	rootCmd.Flags().String(flags.SMTP_FROM_ADDRESS, mail.DefaultSMTPFromAddress, "SMTP-From Address")
	rootCmd.Flags().String(flags.SMTP_FROM_NAME, mail.DefaultSMTPFromName, "SMTP-From Name")
	rootCmd.Flags().String(flags.SMTP_HELO, hostname, "SMTP-HELO/EHLO")
	rootCmd.Flags().String(flags.SMTP_HOST, mail.DefaultSMTPHost, "SMTP-Host")
	rootCmd.Flags().String(flags.SMTP_PASSWORD, "", "SMTP-Password")
	rootCmd.Flags().String(flags.SMTP_USERNAME, "", "SMTP-User")
	rootCmd.Flags().StringArray(flags.SMTP_TO_ADDRESSES, []string{}, "List of recipients")

	rootCmd.AddCommand(completionCmd)

	err = rootCmd.Execute()
	if err != nil {
		return fmt.Errorf("failed to execute root cmd: %w", err)
	}

	return nil
}

func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	// Set the base name of the config file, without the file extension.
	v.SetConfigName(defaultConfigFilename)
	v.SetConfigType(defaultConfigExtension)

	// Set as many paths as you like where viper should look for the
	// config file. We are only looking in the current working directory.
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.config/drone-email/")
	v.AddConfigPath("/etc/drone-email/")

	// Attempt to read the config file, gracefully ignoring errors
	// caused by a config file not being found. Return an error
	// if we cannot parse the config file.
	if err := v.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config: %w", err)
		}
	}

	// Bind to environment variables
	// Works great for simple config names, but needs help for names
	// like --favorite-color which we fix in the bindFlags function
	v.AutomaticEnv()

	// Bind the current command's flags to viper
	bindFlags(cmd, v)

	return nil
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			if len(envPrefix) <= 0 {
				_ = v.BindEnv(f.Name, envVarSuffix)
			} else {
				_ = v.BindEnv(f.Name, fmt.Sprintf("%s_%s", envPrefix, envVarSuffix))
			}
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			_ = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

func newHTMLTemplateVarsByCommand(cmd *cobra.Command) (*mail.CIVars, error) {
	build, err := newBuildByCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize new build struct: %w", err)
	}

	commit, err := newCommitByCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize new commit struct: %w", err)
	}

	deployTo, err := cmd.Flags().GetString(flags.DRONE_DEPLOY_TO)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_DEPLOY_TO, err)
	}

	job, err := newJobByCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize new job struct: %w", err)
	}

	prev, err := newPrevByCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize new prev struct: %w", err)
	}

	pullRequests, err := cmd.Flags().GetInt(flags.DRONE_PULL_REQUEST)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_PULL_REQUEST, err)
	}

	remote, err := newRemoteByCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize new remote struct: %w", err)
	}

	repo, err := newRepoByCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize new repo struct: %w", err)
	}

	tag, err := cmd.Flags().GetString(flags.DRONE_TAG)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_TAG, err)
	}

	yaml, err := newYAMLByCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize new yaml struct: %w", err)
	}

	return &mail.CIVars{
		Build:       build,
		Commit:      commit,
		DeployTo:    deployTo,
		Job:         job,
		Prev:        prev,
		PullRequest: pullRequests,
		Remote:      remote,
		Repo:        repo,
		Tag:         tag,
		Yaml:        yaml,
	}, nil
}

func newBuildByCommand(cmd *cobra.Command) (*domain.Build, error) {
	buildCreated, err := cmd.Flags().GetInt64(flags.DRONE_BUILD_CREATED)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_BUILD_CREATED, err)
	}

	buildEvent, err := cmd.Flags().GetString(flags.DRONE_BUILD_EVENT)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_BUILD_EVENT, err)
	}

	buildFinished, err := cmd.Flags().GetInt64(flags.DRONE_BUILD_FINISHED)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_BUILD_FINISHED, err)
	}

	buildLink, err := cmd.Flags().GetString(flags.DRONE_BUILD_LINK)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_BUILD_LINK, err)
	}

	buildNumber, err := cmd.Flags().GetInt(flags.DRONE_BUILD_NUMBER)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_BUILD_NUMBER, err)
	}

	buildStared, err := cmd.Flags().GetInt64(flags.DRONE_BUILD_STARTED)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_BUILD_STARTED, err)
	}

	buildStatus, err := cmd.Flags().GetString(flags.DRONE_BUILD_STATUS)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_BUILD_STATUS, err)
	}

	build := &domain.Build{
		Created:  buildCreated,
		Event:    buildEvent,
		Finished: buildFinished,
		Link:     buildLink,
		Number:   buildNumber,
		Started:  buildStared,
		Status:   buildStatus,
	}

	return build, nil
}

func newCommitByCommand(cmd *cobra.Command) (*domain.Commit, error) {
	authorAvatar, err := cmd.Flags().GetString(flags.DRONE_COMMIT_AUTHOR_AVATAR)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_COMMIT_AUTHOR_AVATAR, err)
	}

	authorEmail, err := cmd.Flags().GetString(flags.DRONE_COMMIT_AUTHOR_EMAIL)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_COMMIT_AUTHOR_EMAIL, err)
	}

	authorName, err := cmd.Flags().GetString(flags.DRONE_COMMIT_AUTHOR_NAME)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_COMMIT_AUTHOR_NAME, err)
	}

	branch, err := cmd.Flags().GetString(flags.DRONE_COMMIT_BRANCH)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_COMMIT_BRANCH, err)
	}

	link, err := cmd.Flags().GetString(flags.DRONE_COMMIT_LINK)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_COMMIT_LINK, err)
	}

	message, err := cmd.Flags().GetString(flags.DRONE_COMMIT_MESSAGE)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_COMMIT_MESSAGE, err)
	}

	ref, err := cmd.Flags().GetString(flags.DRONE_COMMIT_REF)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_COMMIT_REF, err)
	}

	sha, err := cmd.Flags().GetString(flags.DRONE_COMMIT_SHA)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_COMMIT_SHA, err)
	}

	commit := &domain.Commit{
		Author: &domain.Author{
			Avatar: authorAvatar,
			Email:  authorEmail,
			Name:   authorName,
		},
		Branch:  branch,
		Link:    link,
		Message: message,
		Ref:     ref,
		Sha:     sha,
	}

	return commit, nil
}

func newJobByCommand(cmd *cobra.Command) (*domain.Job, error) {
	exitCode, err := cmd.Flags().GetInt(flags.DRONE_JOB_EXIT_CODE)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_JOB_EXIT_CODE, err)
	}

	finished, err := cmd.Flags().GetInt64(flags.DRONE_BUILD_FINISHED)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_BUILD_FINISHED, err)
	}

	started, err := cmd.Flags().GetInt64(flags.DRONE_BUILD_STARTED)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_BUILD_STARTED, err)
	}

	status, err := cmd.Flags().GetString(flags.DRONE_JOB_STATUS)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_JOB_STATUS, err)
	}

	job := &domain.Job{
		ExitCode: exitCode,
		Finished: finished,
		Started:  started,
		Status:   status,
	}

	return job, nil
}

func newPrevByCommand(cmd *cobra.Command) (*domain.Prev, error) {
	prevBuildNumber, err := cmd.Flags().GetInt(flags.DRONE_PREV_BUILD_NUMBER)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_PREV_BUILD_NUMBER, err)
	}

	prevBuildStatus, err := cmd.Flags().GetString(flags.DRONE_PREV_BUILD_STATUS)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_PREV_BUILD_STATUS, err)
	}

	prevCommitSha, err := cmd.Flags().GetString(flags.DRONE_PREV_COMMIT_SHA)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_PREV_COMMIT_SHA, err)
	}

	prev := &domain.Prev{
		Build: &domain.PrevBuild{
			Number: prevBuildNumber,
			Status: prevBuildStatus,
		},
		Commit: &domain.PrevCommit{
			Sha: prevCommitSha,
		},
	}

	return prev, nil
}

func newRemoteByCommand(cmd *cobra.Command) (*domain.Remote, error) {
	remoteURL, err := cmd.Flags().GetString(flags.DRONE_REMOTE_URL)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_REMOTE_URL, err)
	}

	remote := &domain.Remote{
		URL: remoteURL,
	}

	return remote, nil
}

func newRepoByCommand(cmd *cobra.Command) (*domain.Repo, error) {
	avatar, err := cmd.Flags().GetString(flags.DRONE_REPO_AVATAR)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_REPO_AVATAR, err)
	}

	branch, err := cmd.Flags().GetString(flags.DRONE_REPO_BRANCH)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_REPO_BRANCH, err)
	}

	fullName, err := cmd.Flags().GetString(flags.DRONE_REPO_NAME)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_REPO_NAME, err)
	}

	link, err := cmd.Flags().GetString(flags.DRONE_REPO_LINK)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_REPO_LINK, err)
	}

	name, err := cmd.Flags().GetString(flags.DRONE_REPO)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_REPO, err)
	}

	owner, err := cmd.Flags().GetString(flags.DRONE_REPO_OWNER)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_REPO_OWNER, err)
	}

	private, err := cmd.Flags().GetBool(flags.DRONE_REPO_PRIVATE)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_REPO_PRIVATE, err)
	}

	scm, err := cmd.Flags().GetString(flags.DRONE_REPO_SCM)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_REPO_SCM, err)
	}

	trusted, err := cmd.Flags().GetBool(flags.DRONE_REPO_TRUSTED)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_REPO_TRUSTED, err)
	}

	remote := &domain.Repo{
		Avatar:   avatar,
		Branch:   branch,
		FullName: fullName,
		Link:     link,
		Name:     name,
		Owner:    owner,
		Private:  private,
		SCM:      scm,
		Trusted:  trusted,
	}

	return remote, nil
}

func newYAMLByCommand(cmd *cobra.Command) (*domain.Yaml, error) {
	signed, err := cmd.Flags().GetBool(flags.DRONE_YAML_SIGNED)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_YAML_SIGNED, err)
	}

	verified, err := cmd.Flags().GetBool(flags.DRONE_YAML_VERIFIED)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.DRONE_YAML_VERIFIED, err)
	}

	yaml := &domain.Yaml{
		Signed:   signed,
		Verified: verified,
	}

	return yaml, nil
}

func newSMTPSettingsByCommand(cmd *cobra.Command) (*domain.SMTPSettings, error) {
	smtpStartTLS, err := cmd.Flags().GetBool(flags.SMTP_START_TLS)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.SMTP_START_TLS, err)
	}

	smtpFromAddress, err := cmd.Flags().GetString(flags.SMTP_FROM_ADDRESS)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.SMTP_FROM_ADDRESS, err)
	}

	smtpFromName, err := cmd.Flags().GetString(flags.SMTP_FROM_NAME)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.SMTP_FROM_NAME, err)
	}

	smtpHELOName, err := cmd.Flags().GetString(flags.SMTP_HELO)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.SMTP_HELO, err)
	}

	smtpHost, err := cmd.Flags().GetString(flags.SMTP_HOST)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.SMTP_HOST, err)
	}

	smtpPassword, err := cmd.Flags().GetString(flags.SMTP_PASSWORD)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.SMTP_PASSWORD, err)
	}

	smtpPort, err := cmd.Flags().GetInt(flags.SMTP_PORT)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.SMTP_PORT, err)
	}

	smtpTLSInsecureSkipVerify, err := cmd.Flags().GetBool(flags.SMTP_TLS_INSECURE_SKIP_VERIFY)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.SMTP_TLS_INSECURE_SKIP_VERIFY, err)
	}

	smtpUsername, err := cmd.Flags().GetString(flags.SMTP_USERNAME)
	if err != nil {
		return nil, fmt.Errorf("failed to detect value of %s: %w", flags.SMTP_USERNAME, err)
	}

	return &domain.SMTPSettings{
		FromAddress:           smtpFromAddress,
		FromName:              smtpFromName,
		HELOName:              smtpHELOName,
		Host:                  smtpHost,
		Password:              smtpPassword,
		Port:                  smtpPort,
		StartTLS:              smtpStartTLS,
		TLSInsecureSkipVerify: smtpTLSInsecureSkipVerify,
		Username:              smtpUsername,
	}, nil
}
