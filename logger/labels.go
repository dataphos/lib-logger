package logger

type Labels map[string]string

type L = Labels

// Add adds new and overwrites existing keys.
func (l Labels) Add(labels Labels) Labels {
	for key, val := range labels {
		l[key] = val
	}

	return l
}

// Del deletes keys.
func (l Labels) Del(keys ...string) Labels {
	for _, key := range keys {
		delete(l, key)
	}

	return l
}

// Clone deep copy.
func (l Labels) Clone() Labels {
	clone := Labels{}
	for key, val := range l {
		clone[key] = val
	}

	return clone
}
