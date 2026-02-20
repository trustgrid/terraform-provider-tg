package docs

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContainerStateDocs verifies the container_state documentation quality
func TestContainerStateDocs(t *testing.T) {
	docPath := filepath.Join("resources", "container_state.md")
	content, err := os.ReadFile(docPath)
	require.NoError(t, err, "Documentation file should exist")

	doc := string(content)

	t.Run("HasRequiredFrontmatter", func(t *testing.T) {
		assertFrontmatter(t, doc, "tg_container_state")
	})

	t.Run("HasRequiredSections", func(t *testing.T) {
		assertRequiredSections(t, doc)
	})

	t.Run("HasExampleUsage", func(t *testing.T) {
		assert.Contains(t, doc, "## Example Usage", "Documentation should contain Example Usage section")
	})

	t.Run("HasCommonUseCases", func(t *testing.T) {
		assert.Contains(t, doc, "## Common Use Cases", "Documentation should contain Common Use Cases section")
	})

	t.Run("HasImportSection", func(t *testing.T) {
		assert.Contains(t, doc, "## Import", "Documentation should contain Import section")
	})

	t.Run("HasBehaviorNotes", func(t *testing.T) {
		assert.Contains(t, doc, "## Behavior Notes", "Documentation should contain Behavior Notes section")
	})

	t.Run("HCLExamplesAreValid", func(t *testing.T) {
		examples := extractHCLExamples(doc)
		require.NotEmpty(t, examples, "Documentation should contain HCL examples")

		for i, example := range examples {
			assertValidHCL(t, example, i+1)
		}
	})

	t.Run("SchemaDocumentsRequiredFields", func(t *testing.T) {
		// Verify required fields are documented
		assert.Contains(t, doc, "`container_id`", "Schema should document container_id")
		assert.Contains(t, doc, "`enabled`", "Schema should document enabled")
	})

	t.Run("SchemaDocumentsOptionalFields", func(t *testing.T) {
		// Verify optional fields are documented
		assert.Contains(t, doc, "`node_id`", "Schema should document node_id")
		assert.Contains(t, doc, "`cluster_fqdn`", "Schema should document cluster_fqdn")
	})

	t.Run("ExamplesShowBothNodeAndCluster", func(t *testing.T) {
		// Acceptance criteria: examples show both node_id and cluster_fqdn usage
		assert.Contains(t, doc, "node_id", "Examples should show node_id usage")
		assert.Contains(t, doc, "cluster_fqdn", "Examples should show cluster_fqdn usage")
	})
}

// TestContainerRestartDocs verifies the container_restart documentation quality
func TestContainerRestartDocs(t *testing.T) {
	docPath := filepath.Join("resources", "container_restart.md")
	content, err := os.ReadFile(docPath)
	require.NoError(t, err, "Documentation file should exist")

	doc := string(content)

	t.Run("HasRequiredFrontmatter", func(t *testing.T) {
		assertFrontmatter(t, doc, "tg_container_restart")
	})

	t.Run("HasRequiredSections", func(t *testing.T) {
		assertRequiredSections(t, doc)
	})

	t.Run("HasExampleUsage", func(t *testing.T) {
		assert.Contains(t, doc, "## Example Usage", "Documentation should contain Example Usage section")
	})

	t.Run("HasCommonUseCases", func(t *testing.T) {
		assert.Contains(t, doc, "## Common Use Cases", "Documentation should contain Common Use Cases section")
	})

	t.Run("HasImageUpdateWorkflow", func(t *testing.T) {
		// Acceptance criteria: examples show common use cases (image update workflow)
		assert.Contains(t, doc, "Image Update Workflow", "Documentation should contain Image Update Workflow example")
	})

	t.Run("HasImportSection", func(t *testing.T) {
		assert.Contains(t, doc, "## Import", "Documentation should contain Import section")
	})

	t.Run("HasHowItWorksSection", func(t *testing.T) {
		assert.Contains(t, doc, "## How It Works", "Documentation should explain how the restart works")
	})

	t.Run("HCLExamplesAreValid", func(t *testing.T) {
		examples := extractHCLExamples(doc)
		require.NotEmpty(t, examples, "Documentation should contain HCL examples")

		for i, example := range examples {
			assertValidHCL(t, example, i+1)
		}
	})

	t.Run("SchemaDocumentsRequiredFields", func(t *testing.T) {
		// Verify required fields are documented
		assert.Contains(t, doc, "`container_id`", "Schema should document container_id")
	})

	t.Run("SchemaDocumentsOptionalFields", func(t *testing.T) {
		// Verify optional fields are documented
		assert.Contains(t, doc, "`node_id`", "Schema should document node_id")
		assert.Contains(t, doc, "`cluster_fqdn`", "Schema should document cluster_fqdn")
		assert.Contains(t, doc, "`triggers`", "Schema should document triggers")
	})

	t.Run("ExplainsForceNewBehavior", func(t *testing.T) {
		// The restart behavior is triggered by ForceNew on triggers
		assert.Contains(t, doc, "ForceNew", "Documentation should explain ForceNew behavior on triggers")
	})

	t.Run("ShowsTerraformApplyCommand", func(t *testing.T) {
		// Acceptance criteria: show how to deploy new version
		assert.Contains(t, doc, "terraform apply", "Documentation should show terraform apply command")
	})
}

// TestDocsFollowExistingFormat verifies docs follow the same format as existing docs
func TestDocsFollowExistingFormat(t *testing.T) {
	containerDocPath := filepath.Join("resources", "container.md")
	_, err := os.ReadFile(containerDocPath)
	require.NoError(t, err, "Reference documentation file (container.md) should exist")

	// Check that both new docs follow the standard Terraform documentation structure
	newDocs := []string{"container_state.md", "container_restart.md"}

	for _, docFile := range newDocs {
		t.Run(docFile, func(t *testing.T) {
			docPath := filepath.Join("resources", docFile)
			content, err := os.ReadFile(docPath)
			require.NoError(t, err, "Documentation file should exist: %s", docFile)

			doc := string(content)

			// These are standard Terraform documentation requirements, not conditional
			assert.Contains(t, doc, "page_title:", "Should have page_title in frontmatter")
			assert.Contains(t, doc, "generated by https://github.com/hashicorp/terraform-plugin-docs",
				"Should have terraform-plugin-docs generated comment")
			assert.Contains(t, doc, "<!-- schema generated by tfplugindocs -->",
				"Should have schema generated comment")

			// Check for same heading style (# resource_name (Resource))
			resourceNamePattern := regexp.MustCompile(`# \w+ \(Resource\)`)
			assert.Regexp(t, resourceNamePattern, doc, "Should have resource heading in correct format")
		})
	}
}

// assertFrontmatter checks that the documentation has valid YAML frontmatter
func assertFrontmatter(t *testing.T, doc, resourceName string) {
	t.Helper()

	// Check for frontmatter delimiters
	assert.True(t, strings.HasPrefix(doc, "---"), "Documentation should start with YAML frontmatter delimiter")
	assert.True(t, strings.Count(doc, "---") >= 2, "Documentation should have closing frontmatter delimiter")

	// Check for page_title
	assert.Contains(t, doc, "page_title:", "Frontmatter should contain page_title")
	assert.Contains(t, doc, resourceName, "Frontmatter should reference the resource name")

	// Check for description
	assert.Contains(t, doc, "description:", "Frontmatter should contain description")
}

// assertRequiredSections checks for required Terraform documentation sections
func assertRequiredSections(t *testing.T, doc string) {
	t.Helper()

	requiredSections := []string{
		"## Schema",
		"### Required",
	}

	for _, section := range requiredSections {
		assert.Contains(t, doc, section, "Documentation should contain section: %s", section)
	}
}

// extractHCLExamples extracts all HCL code blocks from markdown
func extractHCLExamples(doc string) []string {
	// Match ```terraform or ```hcl code blocks
	pattern := regexp.MustCompile("```(?:terraform|hcl)\n([\\s\\S]*?)```")
	matches := pattern.FindAllStringSubmatch(doc, -1)

	var examples []string
	for _, match := range matches {
		if len(match) > 1 {
			examples = append(examples, match[1])
		}
	}

	return examples
}

// assertValidHCL checks that an HCL snippet is syntactically valid
func assertValidHCL(t *testing.T, hclCode string, exampleNum int) {
	t.Helper()

	parser := hclparse.NewParser()
	_, diags := parser.ParseHCL([]byte(hclCode), "example.tf")

	// Filter out diagnostics that are just about missing variables or references
	// We only care about syntax errors
	var syntaxErrors hcl.Diagnostics
	for _, diag := range diags {
		if diag.Severity == hcl.DiagError {
			// Skip errors about undefined variables - those are expected in examples
			if !strings.Contains(diag.Summary, "Variables not allowed") &&
				!strings.Contains(diag.Summary, "Unsupported attribute") {
				syntaxErrors = append(syntaxErrors, diag)
			}
		}
	}

	require.Empty(t, syntaxErrors, "HCL example %d should be syntactically valid", exampleNum)
}

// TestDocumentationFilesExist checks that documentation files exist
func TestDocumentationFilesExist(t *testing.T) {
	requiredDocs := []string{
		"resources/container_state.md",
		"resources/container_restart.md",
	}

	for _, docPath := range requiredDocs {
		t.Run(docPath, func(t *testing.T) {
			_, err := os.Stat(docPath)
			assert.NoError(t, err, "Documentation file should exist: %s", docPath)
		})
	}
}

// TestSchemaDescriptionsArePresent verifies that all schema fields have descriptions
func TestSchemaDescriptionsArePresent(t *testing.T) {
	testCases := []struct {
		path         string
		descriptions []string
	}{
		{
			path: "resources/container_state.md",
			descriptions: []string{
				"Container ID",
				"Whether the container should be enabled",
				"Node ID",
				"Cluster FQDN",
			},
		},
		{
			path: "resources/container_restart.md",
			descriptions: []string{
				"Container ID",
				"Node ID",
				"Cluster FQDN",
				"triggers",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			content, err := os.ReadFile(tc.path)
			require.NoError(t, err, "Documentation file should exist: %s", tc.path)

			doc := string(content)

			for _, desc := range tc.descriptions {
				assert.Contains(t, doc, desc, "Documentation should contain description: %s", desc)
			}
		})
	}
}

// TestImportExamplesAreValid checks that import examples show correct format
func TestImportExamplesAreValid(t *testing.T) {
	docs := []struct {
		path           string
		resourceType   string
		expectedFormat string
	}{
		{
			path:           "resources/container_state.md",
			resourceType:   "tg_container_state",
			expectedFormat: "terraform import tg_container_state",
		},
		{
			path:           "resources/container_restart.md",
			resourceType:   "tg_container_restart",
			expectedFormat: "terraform import tg_container_restart",
		},
	}

	for _, tc := range docs {
		t.Run(tc.path, func(t *testing.T) {
			content, err := os.ReadFile(tc.path)
			require.NoError(t, err, "Documentation file should exist: %s", tc.path)

			doc := string(content)

			assert.Contains(t, doc, tc.expectedFormat, "Import section should show correct terraform import command")

			// Check for both node_id and cluster_fqdn import examples
			assert.Contains(t, doc, "Import using node ID", "Import section should show node ID example")
			assert.Contains(t, doc, "Import using cluster FQDN", "Import section should show cluster FQDN example")
		})
	}
}
