package prometheus_data

func get_down_hosts(last []string, cur_up []string) []string {
	down_hosts := make([]string, 0)

	cur_up_map := make(map[string]interface{})
	for _, up := range cur_up {
		cur_up_map[up] = struct{}{}
	}

	for _, host := range last {
		if _, ok := cur_up_map[host]; !ok {
			down_hosts = append(down_hosts, host)
		}
	}

	return down_hosts
}
