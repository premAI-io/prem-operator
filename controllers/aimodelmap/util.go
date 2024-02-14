package aimodelmap

import (
	"fmt"

	a1 "github.com/premAI-io/saas-controller/api/v1alpha1"
	"github.com/premAI-io/saas-controller/controllers/constants"
	v1 "k8s.io/api/core/v1"
)

func FmtConfigMapKey(engineName a1.AIEngineName, variantName string, fieldName constants.AIModelSpecFieldName) string {
	return fmt.Sprintf("%s-%s-%s", engineName, variantName, fieldName)
}

func GetEngineConfigFileData(cm *v1.ConfigMap, engineName a1.AIEngineName, variantName string) string {
	return cm.Data[FmtConfigMapKey(engineName, variantName, constants.AIModelMapSpecEngineConfig)]
}

func SetEngineConfigFileData(cm *v1.ConfigMap, engineName a1.AIEngineName, variantName string, data string) {
	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}
	cm.Data[FmtConfigMapKey(engineName, variantName, constants.AIModelMapSpecEngineConfig)] = data
}
