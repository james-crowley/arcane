package templates

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/getarcaneapp/arcane/cli/internal/client"
	"github.com/getarcaneapp/arcane/cli/internal/cmdutil"
	"github.com/getarcaneapp/arcane/cli/internal/output"
	"github.com/getarcaneapp/arcane/cli/internal/types"
	"github.com/getarcaneapp/arcane/types/base"
	"github.com/getarcaneapp/arcane/types/env"
	"github.com/getarcaneapp/arcane/types/template"
	"github.com/spf13/cobra"
)

var (
	limitFlag  int
	forceFlag  bool
	jsonOutput bool

	templateUpdateName        string
	templateUpdateFile        string
	templateUpdateDescription string
	templateUpdateEnvFile     string
	templateCreateName        string
	templateCreateFile        string
	templateCreateDescription string
	templateCreateEnvFile     string
	templateDownloadOutput    string
	templateDefaultsSaveFile  string
	templateDefaultsEnvFile   string
	templateVarsUpdateFile    string
	templateRegUpdateName     string
	templateRegUpdateURL      string
	templateRegUpdateDesc     string
	templateRegUpdateEnabled  bool
	templateRegUpdateDisabled bool
)

// TemplatesCmd is the parent command for template operations
var TemplatesCmd = &cobra.Command{
	Use:     "templates",
	Aliases: []string{"template", "tpl"},
	Short:   "Manage Docker Compose templates",
}

var listCmd = &cobra.Command{
	Use:          "list",
	Aliases:      []string{"ls"},
	Short:        "List templates",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewFromConfig()
		if err != nil {
			return err
		}

		path := types.Endpoints.Templates()
		effectiveLimit := cmdutil.EffectiveLimit(cmd, "templates", "limit", limitFlag, 20)
		if effectiveLimit > 0 {
			path = fmt.Sprintf("%s?limit=%d", path, effectiveLimit)
		}

		resp, err := c.Get(cmd.Context(), path)
		if err != nil {
			return fmt.Errorf("failed to list templates: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		var result base.ApiResponse[[]template.Template]
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if jsonOutput {
			resultBytes, err := json.MarshalIndent(result.Data, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(resultBytes))
			return nil
		}

		headers := []string{"ID", "NAME", "DESCRIPTION", "CUSTOM", "REMOTE"}
		rows := make([][]string, len(result.Data))
		for i, tpl := range result.Data {
			custom := "no"
			if tpl.IsCustom {
				custom = "yes"
			}
			remote := "no"
			if tpl.IsRemote {
				remote = "yes"
			}
			rows[i] = []string{
				tpl.ID,
				tpl.Name,
				tpl.Description,
				custom,
				remote,
			}
		}

		output.Table(headers, rows)
		fmt.Printf("\nTotal: %d templates\n", len(result.Data))
		return nil
	},
}

var allCmd = &cobra.Command{
	Use:          "all",
	Short:        "List all templates (including remote)",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewFromConfig()
		if err != nil {
			return err
		}

		resp, err := c.Get(cmd.Context(), types.Endpoints.TemplatesAll())
		if err != nil {
			return fmt.Errorf("failed to list all templates: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		var result base.ApiResponse[[]template.Template]
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if jsonOutput {
			resultBytes, err := json.MarshalIndent(result.Data, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(resultBytes))
			return nil
		}

		headers := []string{"ID", "NAME", "DESCRIPTION", "CUSTOM", "REMOTE"}
		rows := make([][]string, len(result.Data))
		for i, tpl := range result.Data {
			custom := "no"
			if tpl.IsCustom {
				custom = "yes"
			}
			remote := "no"
			if tpl.IsRemote {
				remote = "yes"
			}
			rows[i] = []string{
				tpl.ID,
				tpl.Name,
				tpl.Description,
				custom,
				remote,
			}
		}

		output.Table(headers, rows)
		fmt.Printf("\nTotal: %d templates\n", len(result.Data))
		return nil
	},
}

var defaultCmd = &cobra.Command{
	Use:          "default",
	Short:        "Get default templates",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewFromConfig()
		if err != nil {
			return err
		}

		resp, err := c.Get(cmd.Context(), types.Endpoints.TemplatesDefault())
		if err != nil {
			return fmt.Errorf("failed to get default templates: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		var result base.ApiResponse[template.DefaultTemplatesResponse]
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if jsonOutput {
			resultBytes, err := json.MarshalIndent(result.Data, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(resultBytes))
			return nil
		}

		output.Header("Default Templates")
		output.KeyValue("Compose Template", fmt.Sprintf("%d bytes", len(result.Data.ComposeTemplate)))
		output.KeyValue("Env Template", fmt.Sprintf("%d bytes", len(result.Data.EnvTemplate)))
		return nil
	},
}

var contentCmd = &cobra.Command{
	Use:          "content <template-id>",
	Short:        "Get template content",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewFromConfig()
		if err != nil {
			return err
		}

		resp, err := c.Get(cmd.Context(), types.Endpoints.TemplateContent(args[0]))
		if err != nil {
			return fmt.Errorf("failed to get template content: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		var result base.ApiResponse[template.TemplateContent]
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if jsonOutput {
			resultBytes, err := json.MarshalIndent(result.Data, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(resultBytes))
			return nil
		}

		output.Header("Template Content")
		output.KeyValue("Name", result.Data.Template.Name)
		output.KeyValue("Description", result.Data.Template.Description)
		output.KeyValue("Services", fmt.Sprintf("%d", len(result.Data.Services)))
		output.KeyValue("Env Variables", fmt.Sprintf("%d", len(result.Data.EnvVariables)))
		fmt.Println("\n--- Compose Content ---")
		fmt.Println(result.Data.Content)
		if result.Data.EnvContent != "" {
			fmt.Println("\n--- Environment Content ---")
			fmt.Println(result.Data.EnvContent)
		}
		return nil
	},
}

var registriesCmd = &cobra.Command{
	Use:          "registries",
	Aliases:      []string{"reg"},
	Short:        "List template registries",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewFromConfig()
		if err != nil {
			return err
		}

		resp, err := c.Get(cmd.Context(), types.Endpoints.TemplatesRegistries())
		if err != nil {
			return fmt.Errorf("failed to list registries: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		var result base.ApiResponse[[]template.TemplateRegistry]
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if jsonOutput {
			resultBytes, err := json.MarshalIndent(result.Data, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(resultBytes))
			return nil
		}

		headers := []string{"ID", "NAME", "URL", "ENABLED"}
		rows := make([][]string, len(result.Data))
		for i, reg := range result.Data {
			enabled := "no"
			if reg.Enabled {
				enabled = "yes"
			}
			rows[i] = []string{
				reg.ID,
				reg.Name,
				reg.URL,
				enabled,
			}
		}

		output.Table(headers, rows)
		fmt.Printf("\nTotal: %d registries\n", len(result.Data))
		return nil
	},
}

var variablesCmd = &cobra.Command{
	Use:          "variables",
	Aliases:      []string{"vars"},
	Short:        "List template variables",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewFromConfig()
		if err != nil {
			return err
		}

		resp, err := c.Get(cmd.Context(), types.Endpoints.TemplatesVariables())
		if err != nil {
			return fmt.Errorf("failed to list variables: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		var result base.ApiResponse[[]env.Variable]
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if jsonOutput {
			resultBytes, err := json.MarshalIndent(result.Data, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(resultBytes))
			return nil
		}

		headers := []string{"KEY", "VALUE"}
		rows := make([][]string, len(result.Data))
		for i, v := range result.Data {
			rows[i] = []string{
				v.Key,
				v.Value,
			}
		}

		output.Table(headers, rows)
		fmt.Printf("\nTotal: %d variables\n", len(result.Data))
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:          "delete <template-id>",
	Aliases:      []string{"rm", "remove"},
	Short:        "Delete template",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !forceFlag {
			confirmed, err := cmdutil.Confirm(cmd, fmt.Sprintf("Are you sure you want to delete template %s?", args[0]))
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Println("Cancelled")
				return nil
			}
		}

		c, err := client.NewFromConfig()
		if err != nil {
			return err
		}

		resp, err := c.Delete(cmd.Context(), types.Endpoints.Template(args[0]))
		if err != nil {
			return fmt.Errorf("failed to delete template: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if err := cmdutil.EnsureSuccessStatus(resp); err != nil {
			return fmt.Errorf("failed to delete template: %w", err)
		}

		output.Success("Template deleted successfully")
		return nil
	},
}

var deleteRegistryCmd = &cobra.Command{
	Use:          "delete-registry <registry-id>",
	Short:        "Delete template registry",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !forceFlag {
			confirmed, err := cmdutil.Confirm(cmd, fmt.Sprintf("Are you sure you want to delete template registry %s?", args[0]))
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Println("Cancelled")
				return nil
			}
		}

		c, err := client.NewFromConfig()
		if err != nil {
			return err
		}

		resp, err := c.Delete(cmd.Context(), types.Endpoints.TemplateRegistry(args[0]))
		if err != nil {
			return fmt.Errorf("failed to delete registry: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if err := cmdutil.EnsureSuccessStatus(resp); err != nil {
			return fmt.Errorf("failed to delete registry: %w", err)
		}

		output.Success("Template registry deleted successfully")
		return nil
	},
}

var getCmd = &cobra.Command{
	Use:          "get <template-id>",
	Short:        "Get a template by ID",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewFromConfig()
		if err != nil {
			return err
		}

		resp, err := c.Get(cmd.Context(), types.Endpoints.Template(args[0]))
		if err != nil {
			return fmt.Errorf("failed to get template: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		var result base.ApiResponse[template.Template]
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if jsonOutput {
			resultBytes, err := json.MarshalIndent(result.Data, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(resultBytes))
			return nil
		}

		tpl := result.Data
		output.Header("Template")
		output.KeyValue("ID", tpl.ID)
		output.KeyValue("Name", tpl.Name)
		output.KeyValue("Description", tpl.Description)
		custom := "no"
		if tpl.IsCustom {
			custom = "yes"
		}
		remote := "no"
		if tpl.IsRemote {
			remote = "yes"
		}
		output.KeyValue("Custom", custom)
		output.KeyValue("Remote", remote)
		return nil
	},
}

var createCmd = &cobra.Command{
	Use:          "create",
	Short:        "Create a new template",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		content, err := os.ReadFile(templateCreateFile)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", templateCreateFile, err)
		}

		req := template.CreateRequest{
			Name:        templateCreateName,
			Description: templateCreateDescription,
			Content:     string(content),
		}

		if templateCreateEnvFile != "" {
			envContent, err := os.ReadFile(templateCreateEnvFile)
			if err != nil {
				return fmt.Errorf("failed to read env file %s: %w", templateCreateEnvFile, err)
			}
			req.EnvContent = string(envContent)
		}

		c, err := client.NewFromConfig()
		if err != nil {
			return err
		}

		resp, err := c.Post(cmd.Context(), types.Endpoints.Templates(), req)
		if err != nil {
			return fmt.Errorf("failed to create template: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		var result base.ApiResponse[template.Template]
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if jsonOutput {
			resultBytes, err := json.MarshalIndent(result.Data, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(resultBytes))
			return nil
		}

		output.Success("Template created successfully")
		output.KeyValue("ID", result.Data.ID)
		output.KeyValue("Name", result.Data.Name)
		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:          "update <template-id>",
	Short:        "Update a template",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		req := template.UpdateRequest{
			Name:        templateUpdateName,
			Description: templateUpdateDescription,
		}

		if templateUpdateFile != "" {
			content, err := os.ReadFile(templateUpdateFile)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", templateUpdateFile, err)
			}
			req.Content = string(content)
		}

		if templateUpdateEnvFile != "" {
			envContent, err := os.ReadFile(templateUpdateEnvFile)
			if err != nil {
				return fmt.Errorf("failed to read env file %s: %w", templateUpdateEnvFile, err)
			}
			req.EnvContent = string(envContent)
		}

		c, err := client.NewFromConfig()
		if err != nil {
			return err
		}

		resp, err := c.Put(cmd.Context(), types.Endpoints.Template(args[0]), req)
		if err != nil {
			return fmt.Errorf("failed to update template: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		var result base.ApiResponse[template.Template]
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if jsonOutput {
			resultBytes, err := json.MarshalIndent(result.Data, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(resultBytes))
			return nil
		}

		output.Success("Template updated successfully")
		output.KeyValue("ID", result.Data.ID)
		output.KeyValue("Name", result.Data.Name)
		return nil
	},
}

var downloadCmd = &cobra.Command{
	Use:          "download <template-id>",
	Short:        "Download template compose file",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewFromConfig()
		if err != nil {
			return err
		}

		resp, err := c.Post(cmd.Context(), types.Endpoints.TemplateDownload(args[0]), nil)
		if err != nil {
			return fmt.Errorf("failed to download template: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if err := cmdutil.EnsureSuccessStatus(resp); err != nil {
			return fmt.Errorf("failed to download template: %w", err)
		}

		var result base.ApiResponse[template.TemplateContent]
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if templateDownloadOutput != "" {
			dir := filepath.Dir(templateDownloadOutput)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
			if err := os.WriteFile(templateDownloadOutput, []byte(result.Data.Content), 0o600); err != nil {
				return fmt.Errorf("failed to write file %s: %w", templateDownloadOutput, err)
			}
			output.Success("Template downloaded to %s", templateDownloadOutput)
			return nil
		}

		fmt.Print(result.Data.Content)
		return nil
	},
}

var defaultsSaveCmd = &cobra.Command{
	Use:          "defaults-save",
	Short:        "Save default templates",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		content, err := os.ReadFile(templateDefaultsSaveFile)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", templateDefaultsSaveFile, err)
		}

		req := template.SaveDefaultTemplatesRequest{
			ComposeContent: string(content),
		}

		if templateDefaultsEnvFile != "" {
			envContent, err := os.ReadFile(templateDefaultsEnvFile)
			if err != nil {
				return fmt.Errorf("failed to read env file %s: %w", templateDefaultsEnvFile, err)
			}
			req.EnvContent = string(envContent)
		}

		c, err := client.NewFromConfig()
		if err != nil {
			return err
		}

		resp, err := c.Post(cmd.Context(), types.Endpoints.TemplatesDefault(), req)
		if err != nil {
			return fmt.Errorf("failed to save default templates: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if err := cmdutil.EnsureSuccessStatus(resp); err != nil {
			return fmt.Errorf("failed to save default templates: %w", err)
		}

		if jsonOutput {
			var result base.ApiResponse[template.DefaultTemplatesResponse]
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			resultBytes, err := json.MarshalIndent(result.Data, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(resultBytes))
			return nil
		}

		output.Success("Default templates saved successfully")
		return nil
	},
}

var variablesUpdateCmd = &cobra.Command{
	Use:          "variables-update",
	Aliases:      []string{"vars-update"},
	Short:        "Update template variables",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		content, err := os.ReadFile(templateVarsUpdateFile)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", templateVarsUpdateFile, err)
		}

		var variables []env.Variable
		if err := json.Unmarshal(content, &variables); err != nil {
			return fmt.Errorf("failed to parse variables JSON: %w", err)
		}

		c, err := client.NewFromConfig()
		if err != nil {
			return err
		}

		resp, err := c.Put(cmd.Context(), types.Endpoints.TemplatesVariables(), variables)
		if err != nil {
			return fmt.Errorf("failed to update variables: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if err := cmdutil.EnsureSuccessStatus(resp); err != nil {
			return fmt.Errorf("failed to update variables: %w", err)
		}

		if jsonOutput {
			var result base.ApiResponse[[]env.Variable]
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			resultBytes, err := json.MarshalIndent(result.Data, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(resultBytes))
			return nil
		}

		output.Success("Template variables updated successfully")
		return nil
	},
}

var registriesUpdateCmd = &cobra.Command{
	Use:          "registries-update <registry-id>",
	Short:        "Update a template registry",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		req := template.UpdateRegistryRequest{
			Name:        templateRegUpdateName,
			URL:         templateRegUpdateURL,
			Description: templateRegUpdateDesc,
			Enabled:     templateRegUpdateEnabled && !templateRegUpdateDisabled,
		}

		c, err := client.NewFromConfig()
		if err != nil {
			return err
		}

		resp, err := c.Put(cmd.Context(), types.Endpoints.TemplateRegistry(args[0]), req)
		if err != nil {
			return fmt.Errorf("failed to update registry: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		var result base.ApiResponse[template.TemplateRegistry]
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if jsonOutput {
			resultBytes, err := json.MarshalIndent(result.Data, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(resultBytes))
			return nil
		}

		output.Success("Template registry updated successfully")
		output.KeyValue("ID", result.Data.ID)
		output.KeyValue("Name", result.Data.Name)
		return nil
	},
}

var fetchCmd = &cobra.Command{
	Use:          "fetch",
	Short:        "Fetch remote template registries",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewFromConfig()
		if err != nil {
			return err
		}

		resp, err := c.Get(cmd.Context(), types.Endpoints.TemplateFetch())
		if err != nil {
			return fmt.Errorf("failed to fetch templates: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if err := cmdutil.EnsureSuccessStatus(resp); err != nil {
			return fmt.Errorf("failed to fetch templates: %w", err)
		}

		if jsonOutput {
			var result base.ApiResponse[any]
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			resultBytes, err := json.MarshalIndent(result.Data, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(resultBytes))
			return nil
		}

		output.Success("Remote templates fetched successfully")
		return nil
	},
}

func init() {
	TemplatesCmd.AddCommand(listCmd)
	TemplatesCmd.AddCommand(allCmd)
	TemplatesCmd.AddCommand(defaultCmd)
	TemplatesCmd.AddCommand(contentCmd)
	TemplatesCmd.AddCommand(registriesCmd)
	TemplatesCmd.AddCommand(variablesCmd)
	TemplatesCmd.AddCommand(deleteCmd)
	TemplatesCmd.AddCommand(deleteRegistryCmd)
	TemplatesCmd.AddCommand(getCmd)
	TemplatesCmd.AddCommand(createCmd)
	TemplatesCmd.AddCommand(updateCmd)
	TemplatesCmd.AddCommand(downloadCmd)
	TemplatesCmd.AddCommand(defaultsSaveCmd)
	TemplatesCmd.AddCommand(variablesUpdateCmd)
	TemplatesCmd.AddCommand(registriesUpdateCmd)
	TemplatesCmd.AddCommand(fetchCmd)

	listCmd.Flags().IntVarP(&limitFlag, "limit", "n", 20, "Number of templates to show")
	listCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	allCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	defaultCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	contentCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	registriesCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	variablesCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	deleteCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force deletion without confirmation")
	deleteCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	deleteRegistryCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force deletion without confirmation")
	deleteRegistryCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	// get command flags
	getCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	// create command flags
	createCmd.Flags().StringVar(&templateCreateName, "name", "", "Template name")
	createCmd.Flags().StringVar(&templateCreateFile, "file", "", "Path to Docker Compose file")
	createCmd.Flags().StringVar(&templateCreateDescription, "description", "", "Template description")
	createCmd.Flags().StringVar(&templateCreateEnvFile, "env-file", "", "Path to environment file")
	createCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	_ = createCmd.MarkFlagRequired("name")
	_ = createCmd.MarkFlagRequired("file")

	// update command flags
	updateCmd.Flags().StringVar(&templateUpdateName, "name", "", "Template name")
	updateCmd.Flags().StringVar(&templateUpdateFile, "file", "", "Path to Docker Compose file")
	updateCmd.Flags().StringVar(&templateUpdateDescription, "description", "", "Template description")
	updateCmd.Flags().StringVar(&templateUpdateEnvFile, "env-file", "", "Path to environment file")
	updateCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	// download command flags
	downloadCmd.Flags().StringVarP(&templateDownloadOutput, "output", "o", "", "Output file path")
	downloadCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	// defaults-save command flags
	defaultsSaveCmd.Flags().StringVar(&templateDefaultsSaveFile, "file", "", "Path to compose content file")
	defaultsSaveCmd.Flags().StringVar(&templateDefaultsEnvFile, "env-file", "", "Path to environment content file")
	defaultsSaveCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	_ = defaultsSaveCmd.MarkFlagRequired("file")

	// variables-update command flags
	variablesUpdateCmd.Flags().StringVar(&templateVarsUpdateFile, "file", "", "Path to variables JSON file")
	variablesUpdateCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	_ = variablesUpdateCmd.MarkFlagRequired("file")

	// registries-update command flags
	registriesUpdateCmd.Flags().StringVar(&templateRegUpdateName, "name", "", "Registry name")
	registriesUpdateCmd.Flags().StringVar(&templateRegUpdateURL, "url", "", "Registry URL")
	registriesUpdateCmd.Flags().StringVar(&templateRegUpdateDesc, "description", "", "Registry description")
	registriesUpdateCmd.Flags().BoolVar(&templateRegUpdateEnabled, "enabled", false, "Enable registry")
	registriesUpdateCmd.Flags().BoolVar(&templateRegUpdateDisabled, "disabled", false, "Disable registry")
	registriesUpdateCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	// fetch command flags
	fetchCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
}
