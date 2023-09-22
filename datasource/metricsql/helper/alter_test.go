package helper_test

import (
	"testing"

	"github.com/VictoriaMetrics/metricsql"
	"github.com/chengchung/ServerStatus/datasource/metricsql/helper"
)

func TestAlterExpr(t *testing.T) {
	str, err := helper.AlterExpr("sum_over_time(m_name[0m])-sum_over_time(m_name2[1m])==1", helper.AlterOption{
		RollupOpt: &helper.RollupOption{
			Window: "1m",
		},
		Filters: []metricsql.LabelFilter{
			{Label: "a", Value: "b"},
			{Label: "c", Value: "d"},
		},
	})

	if err != nil {
		t.Error(err)
	}
	if str != `(sum_over_time(m_name{a="b",c="d"}[1m]) - sum_over_time(m_name2{a="b",c="d"}[1m])) == 1` {
		t.Error(str)
	}
}

func TestAlterExpr2(t *testing.T) {
	str, err := helper.AlterExpr(`sum(time() - up_time) by (ab)`, helper.AlterOption{
		RollupOpt: &helper.RollupOption{
			Window: "1m",
		},
		Filters: []metricsql.LabelFilter{
			{Label: "a", Value: "b"},
			{Label: "c", Value: "d"},
		},
		AggrBy: []string{"hostname", "device"},
	})

	if err != nil {
		t.Error(err)
	}
	if str != `sum(time() - up_time{a="b",c="d"}) by(hostname,device)` {
		t.Error(str)
	}
}

func TestAlterExpr3(t *testing.T) {
	expr1, err := metricsql.Parse(`up_time{a="b",c="d"} or up_time{a="e",c="f"}`)
	if err != nil {
		t.Error(err)
	}

	expr2, err := helper.AlterExpr(`sum(b) by (a)`, helper.AlterOption{
		AggrBy:   []string{"hostname", "device"},
		AggrArgs: []metricsql.Expr{expr1},
	})
	if err != nil {
		t.Error(err)
	}
	if expr2 != `sum(up_time{a="b",c="d"} or up_time{a="e",c="f"}) by(hostname,device)` {
		t.Error(expr2)
	}
}

func TestAlterExpr4(t *testing.T) {
	sub, err := metricsql.Parse(`node_network_receive_bytes_total:30m_inc`)
	if err != nil {
		t.Error(err)
	}
	expr, err := helper.AlterExpr(`increase(node_network_receive_bytes_total[0m])`, helper.AlterOption{
		RollupOpt:  &helper.RollupOption{Window: "1m"},
		RollupExpr: sub,
		Filters: []metricsql.LabelFilter{
			{Label: "a", Value: "b"},
			{Label: "c", Value: "d"},
		},
	})

	if err != nil {
		t.Error(err)
	}
	if expr != `increase(node_network_receive_bytes_total:30m_inc{a="b",c="d"}[1m])` {
		t.Error(expr)
	}
}
