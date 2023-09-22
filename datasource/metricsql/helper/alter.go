package helper

import (
	"fmt"

	"github.com/VictoriaMetrics/metricsql"
	"github.com/sirupsen/logrus"
)

const (
	example_expr_pattern = "a[%s]"
)

func apply_rollup_expr(expr *metricsql.RollupExpr, new metricsql.Expr) error {
	expr.Expr = new
	return nil
}

func apply_rollup(expr *metricsql.RollupExpr, opt *RollupOption) error {
	get_rollup_parts := func(opt *RollupOption) (*MetricSqlRollupParts, error) {
		str := fmt.Sprintf(example_expr_pattern, opt.Window)
		expr, err := metricsql.Parse(str)
		if err != nil {
			return nil, err
		}
		if expr, ok := expr.(*metricsql.RollupExpr); ok {
			return &MetricSqlRollupParts{
				Window:      expr.Window,
				Offset:      expr.Offset,
				Step:        expr.Step,
				InheritStep: expr.InheritStep,
				At:          expr.At,
			}, nil
		} else {
			return nil, fmt.Errorf("invalid expr type %T, expected %T", expr, &metricsql.RollupExpr{})
		}
	}

	rollup, err := get_rollup_parts(opt)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	expr.Window = rollup.Window
	return nil
}

func apply_filters(expr *metricsql.MetricExpr, filters []metricsql.LabelFilter) error {
	for idx := range expr.LabelFilterss {
		expr.LabelFilterss[idx] = append(expr.LabelFilterss[idx], filters...)
	}

	return nil
}

func apply_aggr_args(expr *metricsql.AggrFuncExpr, args []metricsql.Expr) error {
	expr.Args = args
	return nil
}

func apply_aggr_by(expr *metricsql.AggrFuncExpr, aggr_by []string) error {
	expr.Modifier.Op = "by"
	expr.Modifier.Args = aggr_by
	return nil
}

func AlterExpr(str string, opt AlterOption) (string, error) {
	expr, err := metricsql.Parse(str)
	if err != nil {
		logrus.Error(err.Error())
		return "", err
	}

	missed_exprs := make([]metricsql.Expr, 0)
	fn := func(expr metricsql.Expr) {
		switch expr := expr.(type) {
		case *metricsql.RollupExpr:
			if opt.RollupExpr != nil {
				err := apply_rollup_expr(expr, opt.RollupExpr)
				if err != nil {
					logrus.Error(err.Error())
					return
				}
				missed_exprs = append(missed_exprs, opt.RollupExpr)
			}
			if opt.RollupOpt != nil {
				err = apply_rollup(expr, opt.RollupOpt)
				if err != nil {
					logrus.Error(err.Error())
					return
				}
			}
		case *metricsql.MetricExpr:
			if len(opt.Filters) > 0 {
				err = apply_filters(expr, opt.Filters)
				if err != nil {
					logrus.Error(err.Error())
					return
				}
			}
		case *metricsql.AggrFuncExpr:
			//	aggr type includes more than `by` op, but it should be enough for this project
			if len(opt.AggrArgs) > 0 {
				err = apply_aggr_args(expr, opt.AggrArgs)
				if err != nil {
					logrus.Error(err.Error())
					return
				}
			}
			if len(opt.AggrBy) > 0 {
				err = apply_aggr_by(expr, opt.AggrBy)
				if err != nil {
					logrus.Error(err.Error())
					return
				}
			}
		}
	}

	metricsql.VisitAll(expr, fn)
	for _, expr := range missed_exprs {
		metricsql.VisitAll(expr, fn)
	}

	if err != nil {
		return "", err
	}

	return string(expr.AppendString(nil)), nil
}
