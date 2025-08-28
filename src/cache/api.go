package cache

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

// CloseCache closes the DB connection if initialized.
func CloseCache() error {
	if LyricsDB != nil {
		LyricsDB.Close()
	}
	return nil
}
