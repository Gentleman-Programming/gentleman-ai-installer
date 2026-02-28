package permissions

import (
	"fmt"
	"os"

	"github.com/gentleman-programming/gentle-ai/internal/agents"
	"github.com/gentleman-programming/gentle-ai/internal/components/filemerge"
	"github.com/gentleman-programming/gentle-ai/internal/model"
)

type InjectionResult struct {
	Changed bool
	Files   []string
}

var permissionsOverlayJSON = []byte("{\n  \"permissions\": {\n    \"defaultMode\": \"ask\",\n    \"deny\": [\n      \"rm -rf /\",\n      \"sudo rm -rf /\",\n      \".env\"\n    ]\n  }\n}\n")

// openCodePermissionsOverlayJSON uses the OpenCode "permission" key with bash/read granularity.
var openCodePermissionsOverlayJSON = []byte("{\n  \"permission\": {\n    \"bash\": {\n      \"*\": \"allow\",\n      \"git commit *\": \"ask\",\n      \"git push *\": \"ask\",\n      \"git push\": \"ask\",\n      \"git push --force *\": \"ask\",\n      \"git rebase *\": \"ask\",\n      \"git reset --hard *\": \"ask\"\n    },\n    \"read\": {\n      \"*\": \"allow\",\n      \"*.env\": \"deny\",\n      \"*.env.*\": \"deny\",\n      \"**/.env\": \"deny\",\n      \"**/.env.*\": \"deny\",\n      \"**/secrets/**\": \"deny\",\n      \"**/credentials.json\": \"deny\"\n    }\n  }\n}\n")

func Inject(homeDir string, adapter agents.Adapter) (InjectionResult, error) {
	settingsPath := adapter.SettingsPath(homeDir)
	if settingsPath == "" {
		return InjectionResult{}, nil
	}

	overlay := permissionsOverlayJSON
	if adapter.Agent() == model.AgentOpenCode {
		overlay = openCodePermissionsOverlayJSON
	}

	writeResult, err := mergeJSONFile(settingsPath, overlay)
	if err != nil {
		return InjectionResult{}, err
	}

	return InjectionResult{Changed: writeResult.Changed, Files: []string{settingsPath}}, nil
}

func mergeJSONFile(path string, overlay []byte) (filemerge.WriteResult, error) {
	baseJSON, err := osReadFile(path)
	if err != nil {
		return filemerge.WriteResult{}, err
	}

	merged, err := filemerge.MergeJSONObjects(baseJSON, overlay)
	if err != nil {
		return filemerge.WriteResult{}, err
	}

	return filemerge.WriteFileAtomic(path, merged, 0o644)
}

var osReadFile = func(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read json file %q: %w", path, err)
	}

	return content, nil
}
