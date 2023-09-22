package helper

import "github.com/VictoriaMetrics/metricsql"

type RollupOption struct {
	Window string
}

type MetricSqlRollupParts struct {
	Window      *metricsql.DurationExpr
	Offset      *metricsql.DurationExpr
	Step        *metricsql.DurationExpr
	InheritStep bool
	At          metricsql.Expr
}

type AlterOption struct {
	RollupExpr metricsql.Expr
	RollupOpt  *RollupOption
	Filters    []metricsql.LabelFilter
	AggrArgs   []metricsql.Expr
	AggrBy     []string
}
