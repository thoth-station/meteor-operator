package shower

import "github.com/thoth-station/meteor-operator/controllers/common"

func getSelector(name string) map[string]string {
	return map[string]string{common.SelectorKey: name}
}
