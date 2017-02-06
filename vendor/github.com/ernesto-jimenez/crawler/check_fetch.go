package crawler

// CheckFetchFunc is used to check whether a page should be fetched during the
// crawl or not
type CheckFetchFunc func(*Request) bool

// CheckFetchStack is a stack of CheckFetchFunc types where all have to pass
// for the fetch to happen.
type CheckFetchStack []CheckFetchFunc

// CheckFetch will return true if all funcs in the stack return true. false otherwise.
func (s CheckFetchStack) CheckFetch(req *Request) bool {
	for _, fn := range s {
		if !fn(req) {
			return false
		}
	}
	return true
}
