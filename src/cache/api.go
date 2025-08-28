package cache

type CacheDocResult struct {
	Ref   string `json:"ref"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Get fetches a value from the DB cache (LyricsDB). Returns miss if LyricsDB is not initialized.
func Get(key string) (string, bool) {
	if LyricsDB == nil {
		return "", false
	}
	return LyricsDB.Get(key)
}

// Set stores a value in the DB cache (LyricsDB). Returns an error if LyricsDB is not initialized.
func Set(key, value string) error {
	if LyricsDB == nil {
		return nil
	}
	return LyricsDB.Set(key, value)
}

// SetWithRef stores a value with an optional reference string.
func SetWithRef(key, value, ref string) error {
	if LyricsDB == nil {
		return nil
	}
	return LyricsDB.SetWithRef(key, value, ref)
}

// CloseCache closes the DB connection if initialized.
func CloseCache() error {
	if LyricsDB != nil {
		LyricsDB.Close()
	}
	return nil
}

// GetAll returns all non-expired cache entries as key/value pairs.
func GetAll() ([]CacheDocResult, error) {
	if LyricsDB == nil {
		return []CacheDocResult{}, nil
	}
	docs, err := LyricsDB.GetAll()
	if err != nil {
		return nil, err
	}

	var results []CacheDocResult
	for _, d := range docs {
		results = append(results, CacheDocResult(d))
	}

	return results, nil
}
