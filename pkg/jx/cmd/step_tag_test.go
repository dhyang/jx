package cmd

import (
	"path/filepath"
	"strings"
	"testing"

	"io/ioutil"
	"os"
	"path"

	"github.com/jenkins-x/jx/pkg/gits"
	"github.com/jenkins-x/jx/pkg/tests"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/stretchr/testify/assert"
	"k8s.io/helm/pkg/chartutil"
)

func TestStepTagCharts(t *testing.T) {
	f, err := ioutil.TempDir("", "test-step-tag-charts")
	assert.NoError(t, err)

	testData := path.Join("test_data", "step_tag_project")
	_, err = os.Stat(testData)
	assert.NoError(t, err)

	err = util.CopyDir(testData, f, true)
	assert.NoError(t, err)

	expectedVersion := "1.2.3"
	expectedImageName := "gcr.io/jstrachan/awesome"

	chartsDir := filepath.Join(f, "charts", "mydemo")
	chartFile := filepath.Join(chartsDir, "Chart.yaml")
	valuesFile := filepath.Join(chartsDir, "values.yaml")

	o := StepTagOptions{}
	o.Out = tests.Output()
	o.Flags.ChartsDir = chartsDir
	o.Flags.Version = expectedVersion
	o.Flags.ChartValueRepository = expectedImageName
	o.git = &gits.GitFake{}
	err = o.Run()
	assert.NoError(t, err)

	// root file
	chart, err := chartutil.LoadChartfile(chartFile)
	assert.NoError(t, err, "failed to load file %s", chartFile)

	assert.Equal(t, expectedVersion, chart.Version, "replaced chart version")

	data, err := ioutil.ReadFile(valuesFile)
	assert.NoError(t, err, "failed to load file %s", valuesFile)
	lines := strings.Split(string(data), "\n")

	foundRepo := false
	foundVersion := false
	for _, line := range lines {
		if strings.HasPrefix(line, valuesYamlRepositoryPrefix) {
			value := strings.TrimSpace(strings.TrimPrefix(line, valuesYamlRepositoryPrefix))
			foundRepo = true
			assert.Equal(t, expectedImageName, value, "versions.yaml repository: attribute")
		} else if strings.HasPrefix(line, valuesYamlTagPrefix) {
			foundVersion = true
			value := strings.TrimSpace(strings.TrimPrefix(line, valuesYamlTagPrefix))
			assert.Equal(t, expectedVersion, value, "versions.yaml tag: attribute")
		}
	}

	assert.True(t, foundRepo, "Failed to find tag '%s' in file %s", valuesYamlRepositoryPrefix, valuesFile)
	assert.True(t, foundVersion, "Failed to find tag '%s' in file %s", valuesYamlTagPrefix, valuesFile)
}
