package compress

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/mitchellh/packer/builder/file"
	env "github.com/mitchellh/packer/helper/builder/testing"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template"
)

func TestDetectFilename(t *testing.T) {
	// Test default / fallback with no file extension
	nakedFilename := Config{OutputPath: "test"}
	nakedFilename.detectFromFilename()
	if nakedFilename.Archive != "tar" {
		t.Error("Expected to find tar archive setting")
	}
	if nakedFilename.Algorithm != "pgzip" {
		t.Error("Expected to find pgzip algorithm setting")
	}

	// Test .archive
	zipFilename := Config{OutputPath: "test.zip"}
	zipFilename.detectFromFilename()
	if zipFilename.Archive != "zip" {
		t.Error("Expected to find zip archive setting")
	}
	if zipFilename.Algorithm != "" {
		t.Error("Expected to find empty algorithm setting")
	}

	// Test .compress
	lz4Filename := Config{OutputPath: "test.lz4"}
	lz4Filename.detectFromFilename()
	if lz4Filename.Archive != "" {
		t.Error("Expected to find empty archive setting")
	}
	if lz4Filename.Algorithm != "lz4" {
		t.Error("Expected to find lz4 algorithm setting")
	}

	// Test .archive.compress with some.extra.dots...
	lotsOfDots := Config{OutputPath: "test.blah.bloo.blee.tar.lz4"}
	lotsOfDots.detectFromFilename()
	if lotsOfDots.Archive != "tar" {
		t.Error("Expected to find tar archive setting")
	}
	if lotsOfDots.Algorithm != "lz4" {
		t.Error("Expected to find lz4 algorithm setting")
	}
}

const simpleTestCase = `
{
    "post-processors": [
        {
            "type": "compress",
            "output": "package.tar.gz"
        }
    ]
}
`

func TestSimpleCompress(t *testing.T) {
	artifact := testArchive(t, simpleTestCase)
	defer artifact.Destroy()

	fi, err := os.Stat("package.tar.gz")
	if err != nil {
		t.Errorf("Unable to read archive: %s", err)
	}
	if fi.IsDir() {
		t.Error("Archive should not be a directory")
	}
}

const zipTestCase = `
{
    "post-processors": [
        {
            "type": "compress",
            "output": "package.zip"
        }
    ]
}
`

func TestZipArchive(t *testing.T) {
	artifact := testArchive(t, zipTestCase)
	defer artifact.Destroy()

	// Verify things look good
	_, err := os.Stat("package.zip")
	if err != nil {
		t.Errorf("Unable to read archive: %s", err)
	}
}

const tarTestCase = `
{
    "post-processors": [
        {
            "type": "compress",
            "output": "package.tar"
        }
    ]
}
`

func TestTarArchive(t *testing.T) {
	artifact := testArchive(t, tarTestCase)
	defer artifact.Destroy()

	// Verify things look good
	_, err := os.Stat("package.tar")
	if err != nil {
		t.Errorf("Unable to read archive: %s", err)
	}
}

const optionsTestCase = `
{
    "post-processors": [
        {
            "type": "compress",
            "output": "package.gz",
            "compression_level": 9
        }
    ]
}
`

func TestCompressOptions(t *testing.T) {
	artifact := testArchive(t, optionsTestCase)
	defer artifact.Destroy()

	// Verify things look good
	_, err := os.Stat("package.gz")
	if err != nil {
		t.Errorf("Unable to read archive: %s", err)
	}
}

// Test Helpers

func setup(t *testing.T) (packer.Ui, packer.Artifact, error) {
	// Create fake UI and Cache
	ui := packer.TestUi(t)
	cache := &packer.FileCache{CacheDir: os.TempDir()}

	// Create config for file builder
	const fileConfig = `{"builders":[{"type":"file","target":"package.txt","content":"Hello world!"}]}`
	tpl, err := template.Parse(strings.NewReader(fileConfig))
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to parse setup configuration: %s", err)
	}

	// Prepare the file builder
	builder := file.Builder{}
	warnings, err := builder.Prepare(tpl.Builders["file"].Config)
	if len(warnings) > 0 {
		for _, warn := range warnings {
			return nil, nil, fmt.Errorf("Configuration warning: %s", warn)
		}
	}
	if err != nil {
		return nil, nil, fmt.Errorf("Invalid configuration: %s", err)
	}

	// Run the file builder
	artifact, err := builder.Run(ui, nil, cache)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to build artifact: %s", err)
	}

	return ui, artifact, err
}

func testArchive(t *testing.T, config string) packer.Artifact {
	if os.Getenv(env.TestEnvVar) == "" {
		t.Skip(fmt.Sprintf(
			"Acceptance tests skipped unless env '%s' set", env.TestEnvVar))
	}

	ui, artifact, err := setup(t)
	if err != nil {
		t.Fatalf("Error bootstrapping test: %s", err)
	}
	if artifact != nil {
		defer artifact.Destroy()
	}

	tpl, err := template.Parse(strings.NewReader(config))
	if err != nil {
		t.Fatalf("Unable to parse test config: %s", err)
	}

	compressor := PostProcessor{}
	compressor.Configure(tpl.PostProcessors[0][0].Config)
	artifactOut, _, err := compressor.PostProcess(ui, artifact)
	if err != nil {
		t.Fatalf("Failed to compress artifact: %s", err)
	}

	return artifactOut
}
