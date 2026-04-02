package entity

// EvaluateRules iterates rules in order and returns the ResultStatus of the first
// matching rule. If no rule matches, it returns APPROVED (fail-open by design).
func EvaluateRules(transaction *TransactionMessage, rules []Rule) DecisionStatus {
	for _, rule := range rules {
		if rule.Matches(transaction) {
			return rule.ResultStatus
		}
	}
	return APPROVED
}
