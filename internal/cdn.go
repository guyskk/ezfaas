package internal

import (
	"log"

	"github.com/guyskk/ezfaas/internal/common"
	"github.com/guyskk/ezfaas/internal/tencent"
)

type TencentCDNCacheConfigParams tencent.CDNCacheConfigParams

func DoConfigCdnCacheTencent(params TencentCDNCacheConfigParams) {
	output, err := tencent.UpdateCDNCacheConfig(
		tencent.CDNCacheConfigParams(params))
	if err != nil {
		log.Fatal(err)
	}
	common.LogPrettyJSON(output)
}
