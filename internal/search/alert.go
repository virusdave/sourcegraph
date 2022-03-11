package search

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Alert struct {
	PrometheusType  string
	Title           string
	Description     string
	ProposedQueries []*ProposedQuery
	// The higher the priority the more important is the alert.
	Priority int
}

func MaxPriorityAlert(alerts ...*Alert) (max *Alert) {
	for _, alert := range alerts {
		if alert == nil {
			continue
		}
		if max == nil || alert.Priority > max.Priority {
			max = alert
		}
	}
	return max
}

// MaxAlerter is a simple struct that provides a thread-safe way
// to aggregate a set of alerts, leaving the highest priority alert
type MaxAlerter struct {
	sync.Mutex
	*Alert
}

func (m *MaxAlerter) Add(a *Alert) {
	m.Lock()
	m.Alert = MaxPriorityAlert(m.Alert, a)
	m.Unlock()
}

type ProposedQuery struct {
	Description string
	Query       string
	PatternType query.SearchType
}

func (q *ProposedQuery) QueryString() string {
	if q.Description != "Remove quotes" {
		switch q.PatternType {
		case query.SearchTypeRegex:
			return q.Query + " patternType:regexp"
		case query.SearchTypeLiteral:
			return q.Query + " patternType:literal"
		case query.SearchTypeStructural:
			return q.Query + " patternType:structural"
		default:
			panic("unreachable")
		}
	}
	return q.Query
}

func AlertForCappedAndExpression() *Alert {
	return &Alert{
		PrometheusType: "exceed_and_expression_search_limit",
		Title:          "Too many files to search for expression",
		Description:    "One expression in the query requires a lot of work! This can happen with negated text searches like '-content:', not-expressions, or and-expressions. Try using the '-file:' or '-repo:' filters to narrow your search (like excluding autogenerated files). We're working on improving this experience in https://github.com/sourcegraph/sourcegraph/issues/9824",
	}
}

// AlertForQuery converts errors in the query to search alerts.
func AlertForQuery(queryString string, err error) *Alert {
	if errors.HasType(err, &query.UnsupportedError{}) || errors.HasType(err, &query.ExpectedOperand{}) {
		return &Alert{
			PrometheusType: "unsupported_and_or_query",
			Title:          "Unable To Process Query",
			Description:    `I'm having trouble understanding that query. Your query contains "and" or "or" operators that make me think they apply to filters like "repo:" or "file:". We only support "and" or "or" operators on search patterns for file contents currently. You can help me by putting parentheses around the search pattern.`,
		}
	}
	return &Alert{
		PrometheusType: "generic_invalid_query",
		Title:          "Unable To Process Query",
		Description:    capFirst(err.Error()),
	}
}

func AlertForTimeout(usedTime time.Duration, suggestTime time.Duration, queryString string, patternType query.SearchType) *Alert {
	q, err := query.ParseLiteral(queryString) // Invariant: query is already validated; guard against error anyway.
	if err != nil {
		return &Alert{
			PrometheusType: "timed_out",
			Title:          "Timed out while searching",
			Description:    fmt.Sprintf("We weren't able to find any results in %s. Try adding timeout: with a higher value.", usedTime.Round(time.Second)),
		}
	}

	return &Alert{
		PrometheusType: "timed_out",
		Title:          "Timed out while searching",
		Description:    fmt.Sprintf("We weren't able to find any results in %s.", usedTime.Round(time.Second)),
		ProposedQueries: []*ProposedQuery{
			{
				Description: "query with longer timeout",
				Query:       fmt.Sprintf("timeout:%v %s", suggestTime, query.OmitField(q, query.FieldTimeout)),
				PatternType: patternType,
			},
		},
	}
}

// capFirst capitalizes the first rune in the given string. It can be safely
// used with UTF-8 strings.
func capFirst(s string) string {
	i := 0
	return strings.Map(func(r rune) rune {
		i++
		if i == 1 {
			return unicode.ToTitle(r)
		}
		return r
	}, s)
}

func AlertForStalePermissions() *Alert {
	return &Alert{
		PrometheusType: "no_resolved_repos__stale_permissions",
		Title:          "Permissions syncing in progress",
		Description:    "Permissions are being synced from your code host, please wait for a minute and try again.",
	}
}

func AlertForStructuralSearchNotSet(queryString string) *Alert {
	return &Alert{
		PrometheusType: "structural_search_not_set",
		Title:          "No results",
		Description:    "It looks like you may have meant to run a structural search, but it is not toggled.",
		ProposedQueries: []*ProposedQuery{
			{
				Description: "Activate structural search",
				Query:       queryString,
				PatternType: query.SearchTypeStructural,
			},
		},
	}
}

func AlertForMissingRepoRevs(missingRepoRevs []*RepositoryRevisions) *Alert {
	var description string
	if len(missingRepoRevs) == 1 {
		if len(missingRepoRevs[0].RevSpecs()) == 1 {
			description = fmt.Sprintf("The repository %s matched by your repo: filter could not be searched because it does not contain the revision %q.", missingRepoRevs[0].Repo.Name, missingRepoRevs[0].RevSpecs()[0])
		} else {
			description = fmt.Sprintf("The repository %s matched by your repo: filter could not be searched because it has multiple specified revisions: @%s.", missingRepoRevs[0].Repo.Name, strings.Join(missingRepoRevs[0].RevSpecs(), ","))
		}
	} else {
		sampleSize := 10
		if sampleSize > len(missingRepoRevs) {
			sampleSize = len(missingRepoRevs)
		}
		repoRevs := make([]string, 0, sampleSize)
		for _, r := range missingRepoRevs[:sampleSize] {
			repoRevs = append(repoRevs, string(r.Repo.Name)+"@"+strings.Join(r.RevSpecs(), ","))
		}
		b := strings.Builder{}
		_, _ = fmt.Fprintf(&b, "%d repositories matched by your repo: filter could not be searched because the following revisions do not exist, or differ but were specified for the same repository:", len(missingRepoRevs))
		for _, rr := range repoRevs {
			_, _ = fmt.Fprintf(&b, "\n* %s", rr)
		}
		if sampleSize < len(missingRepoRevs) {
			b.WriteString("\n* ...")
		}
		description = b.String()
	}
	return &Alert{
		PrometheusType: "missing_repo_revs",
		Title:          "Some repositories could not be searched",
		Description:    description,
	}
}

func AlertForMissingDependencyRepoRevs(missingRepoRevs []*RepositoryRevisions) *Alert {
	var description string
	if len(missingRepoRevs) == 1 {
		if len(missingRepoRevs[0].RevSpecs()) == 1 {
			description = fmt.Sprintf("The dependency %s matched by your repo:deps(...) predicate could not be searched because it does not yet contain the revision %q.", missingRepoRevs[0].Repo.Name, missingRepoRevs[0].RevSpecs()[0])
		} else {
			description = fmt.Sprintf("The dependency %s matched by your repo:deps(...) predicate could not be searched because it has multiple missing revisions: @%s.", missingRepoRevs[0].Repo.Name, strings.Join(missingRepoRevs[0].RevSpecs(), ","))
		}
	} else {
		sampleSize := 10
		if sampleSize > len(missingRepoRevs) {
			sampleSize = len(missingRepoRevs)
		}
		repoRevs := make([]string, 0, sampleSize)
		for _, r := range missingRepoRevs[:sampleSize] {
			repoRevs = append(repoRevs, string(r.Repo.Name)+"@"+strings.Join(r.RevSpecs(), ","))
		}
		b := strings.Builder{}
		_, _ = fmt.Fprintf(&b, "%d dependencies matched by your repo:deps(...) predicate could not be searched because the following revisions either don't exist or aren't yet cloned:", len(missingRepoRevs))
		for _, rr := range repoRevs {
			_, _ = fmt.Fprintf(&b, "\n* %s", rr)
		}
		if sampleSize < len(missingRepoRevs) {
			b.WriteString("\n* ...")
		}
		description = b.String()
	}
	return &Alert{
		PrometheusType: "missing_dependency_repo_revs",
		Title:          "Some dependencies could not be searched",
		Description:    description + "\n\nDependency repository revisions are cloned on demand. Try again in a few seconds.",
	}
}

func AlertForInvalidRevision(revision string) *Alert {
	revision = strings.TrimSuffix(revision, "^0")
	return &Alert{
		Title:       "Invalid revision syntax",
		Description: fmt.Sprintf("We don't know how to interpret the revision (%s) you specified. Learn more about the revision syntax in our documentation: https://docs.sourcegraph.com/code_search/reference/queries#repository-revisions.", revision),
	}
}
