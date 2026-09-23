package boot

import (
	"fmt"
	"sort"
	"strings"
)

// StarterModule 模块化的 Starter 定义
//
// 用于声明模块的依赖关系和可选性，支持冲突检测。
type StarterModule struct {
	Name         string   // Starter 名称
	Dependencies []string // 依赖的其他 Starter 名称
	Optional     bool     // 是否为可选依赖
}

// CompositeStarter 创建组合 Starter
//
// 将多个 Starter 组合成一个逻辑单元，便于模块化声明。
// 示例:
//
//	var WebStarter = boot.CompositeStarter(
//	    HTTPStarter,
//	    ValidationStarter,
//	    ExceptionStarter,
//	)
func CompositeStarter(starters ...StarterModule) StarterModule {
	if len(starters) == 0 {
		return StarterModule{}
	}

	name := starters[0].Name
	deps := make([]string, 0)
	optional := false

	for _, s := range starters {
		// 只添加 Dependencies 字段，不重复添加 Name
		deps = append(deps, s.Dependencies...)
		if s.Optional {
			optional = true
		}
	}

	return StarterModule{
		Name:         name,
		Dependencies: deps,
		Optional:     optional,
	}
}

// Conflict 表示两个 Starter 之间的冲突
type Conflict struct {
	StarterA string // 第一个 Starter
	StarterB string // 第二个 Starter
	Reason   string // 冲突原因
}

// String 返回冲突的可读描述信息。
func (c Conflict) String() string {
	return fmt.Sprintf("冲突: %s <-> %s (%s)", c.StarterA, c.StarterB, c.Reason)
}

// DetectConflicts 检测 Starter 之间的冲突
//
// 检测规则:
//  1. 循环依赖检测
//  2. 重复名称检测
//  3. 缺失依赖检测
func DetectConflicts(starters []StarterModule) []Conflict {
	conflicts := make([]Conflict, 0)

	// 检测重复名称
	nameCount := make(map[string]int)
	for _, s := range starters {
		nameCount[s.Name]++
	}
	for name, count := range nameCount {
		if count > 1 {
			conflicts = append(conflicts, Conflict{
				StarterA: name,
				StarterB: name,
				Reason:   fmt.Sprintf("重复的 Starter 名称 (%d 次)", count),
			})
		}
	}

	// 检测缺失依赖
	nameSet := make(map[string]bool)
	for _, s := range starters {
		nameSet[s.Name] = true
	}
	for _, s := range starters {
		for _, dep := range s.Dependencies {
			if !nameSet[dep] {
				conflicts = append(conflicts, Conflict{
					StarterA: s.Name,
					StarterB: dep,
					Reason:   "依赖的 Starter 不存在",
				})
			}
		}
	}

	// 检测循环依赖
	cycles := detectCycles(starters)
	for _, cycle := range cycles {
		conflicts = append(conflicts, Conflict{
			StarterA: cycle[0],
			StarterB: cycle[1],
			Reason:   fmt.Sprintf("循环依赖: %s", strings.Join(cycle, " -> ")),
		})
	}

	return conflicts
}

// detectCycles 检测循环依赖
func detectCycles(starters []StarterModule) [][]string {
	detector := &cycleDetector{
		adj:      buildDependencyAdjacency(starters),
		visited:  make(map[string]bool),
		recStack: make(map[string]bool),
	}

	for _, s := range starters {
		if !detector.visited[s.Name] {
			detector.dfs(s.Name, []string{})
		}
	}

	return detector.cycles
}

// cycleDetector 循环依赖检测器，封装 DFS 遍历状态。
type cycleDetector struct {
	adj      map[string][]string
	visited  map[string]bool
	recStack map[string]bool
	cycles   [][]string
}

// buildDependencyAdjacency 构建启动器名称到依赖列表的邻接表。
func buildDependencyAdjacency(starters []StarterModule) map[string][]string {
	adj := make(map[string][]string)
	for _, s := range starters {
		adj[s.Name] = s.Dependencies
	}
	return adj
}

// dfs 深度优先遍历依赖图，检测并收集循环依赖。
func (d *cycleDetector) dfs(node string, path []string) {
	d.visited[node] = true
	d.recStack[node] = true
	pathLen := len(path)
	path = append(path, node)

	for _, neighbor := range d.adj[node] {
		if _, exists := d.adj[neighbor]; !exists {
			continue
		}
		if !d.visited[neighbor] {
			d.dfs(neighbor, path)
		} else if d.recStack[neighbor] {
			cycleStart := -1
			for i := pathLen; i < len(path); i++ {
				if path[i] == neighbor {
					cycleStart = i
					break
				}
			}
			if cycleStart == -1 {
				for i := 0; i < pathLen; i++ {
					if path[i] == neighbor {
						cycleStart = i
						break
					}
				}
			}
			if cycleStart >= 0 {
				cycle := make([]string, 0, len(path)-cycleStart+1)
				cycle = append(cycle, path[cycleStart:]...)
				cycle = append(cycle, neighbor)
				d.cycles = append(d.cycles, cycle)
			}
		}
	}

	d.recStack[node] = false
}

// ValidateStarters 验证 Starter 列表的有效性
//
// 返回错误信息（如果有冲突）或 nil（如果有效）。
func ValidateStarters(starters []StarterModule) error {
	conflicts := DetectConflicts(starters)
	if len(conflicts) == 0 {
		return nil
	}

	msgs := make([]string, 0, len(conflicts))
	for _, c := range conflicts {
		msgs = append(msgs, c.String())
	}

	return fmt.Errorf("Starter 验证失败:\n%s", strings.Join(msgs, "\n"))
}

// ResolveDependencies 解析并排序 Starter 依赖
//
// 返回按依赖关系排序的 Starter 列表。
func ResolveDependencies(starters []StarterModule) ([]StarterModule, error) {
	if err := ValidateStarters(starters); err != nil {
		return nil, fmt.Errorf("validate starters: %w", err)
	}

	nameMap := buildModuleNameMap(starters)

	inDegree, adj := buildModuleDependencyGraph(starters, nameMap)

	// 拓扑排序（Kahn 算法），入度为 0 的节点依次出队
	resolved := topoSortModules(nameMap, inDegree, adj)

	if len(resolved) != len(starters) {
		return nil, fmt.Errorf("存在循环依赖")
	}

	return resolved, nil
}

// buildModuleNameMap 构建 Starter 模块名称到实例的映射。
func buildModuleNameMap(starters []StarterModule) map[string]StarterModule {
	nameMap := make(map[string]StarterModule)
	for _, s := range starters {
		nameMap[s.Name] = s
	}
	return nameMap
}

// buildModuleDependencyGraph 构建 Starter 模块依赖图的入度表和邻接表。
func buildModuleDependencyGraph(starters []StarterModule, nameMap map[string]StarterModule) (map[string]int, map[string][]string) {
	inDegree := make(map[string]int)
	adj := make(map[string][]string)

	for _, s := range starters {
		if _, ok := inDegree[s.Name]; !ok {
			inDegree[s.Name] = 0
		}
		for _, dep := range s.Dependencies {
			if _, ok := nameMap[dep]; ok {
				adj[dep] = append(adj[dep], s.Name)
				inDegree[s.Name]++
			}
		}
	}
	return inDegree, adj
}

// topoSortModules 使用 Kahn 算法对 Starter 模块进行拓扑排序。
func topoSortModules(nameMap map[string]StarterModule, inDegree map[string]int, adj map[string][]string) []StarterModule {
	queue := make([]string, 0)
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}

	sort.Strings(queue)

	resolved := make([]StarterModule, 0)
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		resolved = append(resolved, nameMap[name])

		nextDeps := adj[name]
		sort.Strings(nextDeps)
		for _, dep := range nextDeps {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	return resolved
}
