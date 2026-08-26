package aop

// Compose 组合多个切点（AND 逻辑）
//
// 只有当所有切点都匹配时，才认为匹配。
//
// 参数:
//   - pointcuts: 切点列表
//
// 返回值:
//   - PointCut: 组合后的切点
//
// 示例:
//
//	// 匹配 Service 类中以 Get 开头的方法
//	aop.Compose(
//	    aop.MatchByClassName("*Service"),
//	    aop.MatchByNamePrefix("Get"),
//	)
func Compose(pointcuts ...PointCut) PointCut {
	return &compositePointCut{
		pointcuts: pointcuts,
		and:       true,
	}
}

// ComposeOr 组合多个切点（OR 逻辑）
//
// 只要有一个切点匹配，就认为匹配。
//
// 参数:
//   - pointcuts: 切点列表
//
// 返回值:
//   - PointCut: 组合后的切点
//
// 示例:
//
//	// 匹配 GetUser 或 UpdateUser 方法
//	aop.ComposeOr(
//	    aop.MatchByName("GetUser"),
//	    aop.MatchByName("UpdateUser"),
//	)
func ComposeOr(pointcuts ...PointCut) PointCut {
	return &compositePointCut{
		pointcuts: pointcuts,
		and:       false,
	}
}
