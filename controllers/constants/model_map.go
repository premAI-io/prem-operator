package constants

type AIModelSpecFieldName string

const (
	AIModelMapSpecEngineConfig AIModelSpecFieldName = "engineConfigFile"
	AIModelMapSpecModelFiles   AIModelSpecFieldName = "modelFiles"

	AIModelMapDefaultAnnotationKey = "mlcontroller.premlabs.io/aimodelmap"
	AIModelMapDefaultLabelKey      = "mlcontroller.premlabs.io/aimodelmap"
)
