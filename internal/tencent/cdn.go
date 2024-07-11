package tencent

import (
	"strings"

	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

type CDNCacheConfigParams struct {
	Region     string
	Domain     string
	UsageLimit string
}

var (
	ON  string = "on"
	OFF string = "off"
)

func strRef(x string) *string {
	return &x
}

func int64Ref(x int64) *int64 {
	return &x
}

func uint64Ref(x uint64) *uint64 {
	return &x
}

func getNodeCacheRules() []*cdn.RuleCache {
	indexCacheConfig := cdn.RuleCacheConfig{
		Cache: &cdn.CacheConfigCache{
			Switch:             &ON,
			CacheTime:          int64Ref(10),
			CompareMaxAge:      &OFF,
			IgnoreCacheControl: &OFF,
			IgnoreSetCookie:    &OFF,
		},
	}
	staticCacheConfig := cdn.RuleCacheConfig{
		Cache: &cdn.CacheConfigCache{
			Switch:             &ON,
			CacheTime:          int64Ref(10 * 60 * 60),
			CompareMaxAge:      &OFF,
			IgnoreCacheControl: &OFF,
			IgnoreSetCookie:    &OFF,
		},
	}
	apiCacheConfig := cdn.RuleCacheConfig{
		NoCache: &cdn.CacheConfigNoCache{
			Switch: &ON,
		},
	}
	return []*cdn.RuleCache{
		{
			RuleType: strRef("all"),
			RulePaths: []*string{
				strRef("*"),
			},
			CacheConfig: &indexCacheConfig,
		},
		{
			RuleType: strRef("directory"),
			RulePaths: []*string{
				strRef("/api"),
			},
			CacheConfig: &apiCacheConfig,
		},
		{
			RuleType: strRef("directory"),
			RulePaths: []*string{
				strRef("/js"),
				strRef("/css"),
				strRef("/fonts"),
				strRef("/imgs"),
				strRef("/img"),
				strRef("/libs"),
				strRef("/static"),
				strRef("/assets"),
			},
			CacheConfig: &staticCacheConfig,
		},
		{
			RuleType: strRef("path"),
			RulePaths: []*string{
				strRef("/favicon.ico"),
			},
			CacheConfig: &staticCacheConfig,
		},
		{
			RuleType: strRef("path"),
			RulePaths: []*string{
				strRef("/manifest.json"),
				strRef("/service-worker.js"),
			},
			CacheConfig: &indexCacheConfig,
		},
		{
			RuleType: strRef("index"),
			RulePaths: []*string{
				strRef("/"),
			},
			CacheConfig: &indexCacheConfig,
		},
	}
}

func getBrowserCacheRules() []*cdn.MaxAgeRule {
	indexMaxAgeTime := int64Ref(30)
	staticMaxAgeTime := int64Ref(10 * 24 * 60 * 60)
	return []*cdn.MaxAgeRule{
		{
			MaxAgeType: strRef("all"),
			MaxAgeContents: []*string{
				strRef("*"),
			},
			MaxAgeTime:   indexMaxAgeTime,
			FollowOrigin: &ON,
		},
		{
			MaxAgeType: strRef("directory"),
			MaxAgeContents: []*string{
				strRef("/api"),
			},
			MaxAgeTime:   int64Ref(0),
			FollowOrigin: &ON,
		},
		{
			MaxAgeType: strRef("directory"),
			MaxAgeContents: []*string{
				strRef("/js"),
				strRef("/css"),
				strRef("/fonts"),
				strRef("/imgs"),
				strRef("/img"),
				strRef("/libs"),
				strRef("/static"),
				strRef("/assets"),
			},
			MaxAgeTime: staticMaxAgeTime,
		},
		{
			MaxAgeType: strRef("path"),
			MaxAgeContents: []*string{
				strRef("/favicon.ico"),
			},
			MaxAgeTime: staticMaxAgeTime,
		},
		{
			MaxAgeType: strRef("path"),
			MaxAgeContents: []*string{
				strRef("/manifest.json"),
				strRef("/service-worker.js"),
			},
			MaxAgeTime: indexMaxAgeTime,
		},
		{
			MaxAgeType: strRef("index"),
			MaxAgeContents: []*string{
				strRef("/"),
			},
			MaxAgeTime: indexMaxAgeTime,
		},
	}
}

func addUsageLimitRule(request *cdn.UpdateDomainConfigRequest, isEnable bool) {
	status := OFF
	if isEnable {
		status = ON
	}
	request.IpFreqLimit = &cdn.IpFreqLimit{
		Switch: &status,
		Qps:    int64Ref(50),
	}
	request.DownstreamCapping = &cdn.DownstreamCapping{
		Switch: &status,
		CappingRules: []*cdn.CappingRule{
			{
				RuleType: strRef("all"),
				RulePaths: []*string{
					strRef("*"),
				},
				KBpsThreshold: int64Ref(1024),
			},
		},
	}
	bandwidthAlertItem := &cdn.StatisticItem{
		Switch:          &status,
		AlertSwitch:     &ON,
		Type:            strRef("moment"),
		Metric:          strRef("bandwidth"),
		BpsThreshold:    uint64Ref(20 * 1000 * 1000),
		CounterMeasure:  strRef("RETURN_404"),
		Cycle:           uint64Ref(5),
		UnBlockTime:     uint64Ref(60),
		AlertPercentage: uint64Ref(50),
	}
	request.BandwidthAlert = &cdn.BandwidthAlert{
		StatisticItems: []*cdn.StatisticItem{bandwidthAlertItem},
	}
}

func UpdateCDNCacheConfig(
	params CDNCacheConfigParams,
) (*cdn.UpdateDomainConfigResponse, error) {
	provider := common.DefaultProfileProvider()
	credentail, err := provider.GetCredential()
	if err != nil {
		return nil, err
	}
	clientProfile := profile.NewClientProfile()
	client, err := cdn.NewClient(credentail, params.Region, clientProfile)
	if err != nil {
		return nil, err
	}
	request := cdn.NewUpdateDomainConfigRequest()
	request.Domain = &params.Domain
	request.Cache = &cdn.Cache{
		RuleCache: getNodeCacheRules(),
	}
	request.MaxAge = &cdn.MaxAge{
		Switch:      &ON,
		MaxAgeRules: getBrowserCacheRules(),
	}
	// 配置限流和用量封顶
	usageLimit := strings.ToLower(params.UsageLimit)
	if usageLimit == ON {
		addUsageLimitRule(request, true)
	} else if usageLimit == OFF {
		addUsageLimitRule(request, false)
	}
	response, err := client.UpdateDomainConfig(request)
	if err != nil {
		return nil, err
	}
	return response, nil
}
