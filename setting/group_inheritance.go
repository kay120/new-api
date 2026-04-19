package setting

import (
	"encoding/json"
	"sync"

	"github.com/QuantumNous/new-api/common"
)

// groupInheritance 保存"分组继承链"配置，用于渠道路由 fallback。
//
// 语义：如果用户分组 = X，调用时发现 X 分组下没有满足 model 的渠道，
// 按 groupInheritance[X] 列出的顺序依次尝试下一个分组，找到第一个有渠道的为止。
//
// 示例：{"vip": ["default"], "svip": ["vip", "default"]}
//   - vip 用户：先查 vip，查不到 fallback 到 default
//   - svip 用户：先查 svip，再 vip，最后 default
//
// 注意：这只影响"路由选渠道"，不影响计费倍率。倍率仍按用户的原分组 / token 覆写算。
var groupInheritance = map[string][]string{}
var groupInheritanceMutex sync.RWMutex

// GetGroupInheritanceChain 返回分组的 fallback 链（不含自己）。
// 未配置时返回 nil。
func GetGroupInheritanceChain(groupName string) []string {
	groupInheritanceMutex.RLock()
	defer groupInheritanceMutex.RUnlock()
	chain, ok := groupInheritance[groupName]
	if !ok || len(chain) == 0 {
		return nil
	}
	// 返回副本避免调用方误改共享状态
	out := make([]string, len(chain))
	copy(out, chain)
	return out
}

// GetGroupInheritanceCopy 返回整份配置副本（前端读取用）。
func GetGroupInheritanceCopy() map[string][]string {
	groupInheritanceMutex.RLock()
	defer groupInheritanceMutex.RUnlock()
	out := make(map[string][]string, len(groupInheritance))
	for k, v := range groupInheritance {
		vv := make([]string, len(v))
		copy(vv, v)
		out[k] = vv
	}
	return out
}

func GroupInheritance2JSONString() string {
	groupInheritanceMutex.RLock()
	defer groupInheritanceMutex.RUnlock()
	b, err := json.Marshal(groupInheritance)
	if err != nil {
		common.SysLog("error marshalling group inheritance: " + err.Error())
		return "{}"
	}
	return string(b)
}

func UpdateGroupInheritanceByJSONString(jsonStr string) error {
	groupInheritanceMutex.Lock()
	defer groupInheritanceMutex.Unlock()
	if jsonStr == "" {
		groupInheritance = map[string][]string{}
		return nil
	}
	next := map[string][]string{}
	if err := json.Unmarshal([]byte(jsonStr), &next); err != nil {
		return err
	}
	// 清理自继承 / 空项，防止死循环
	for k, chain := range next {
		cleaned := make([]string, 0, len(chain))
		seen := map[string]bool{k: true}
		for _, g := range chain {
			if g == "" || seen[g] {
				continue
			}
			seen[g] = true
			cleaned = append(cleaned, g)
		}
		next[k] = cleaned
	}
	groupInheritance = next
	return nil
}
