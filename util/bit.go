package util

type Bit uint32

func (m *Bit) Set(flag Bit)      { *m = *m | flag }
func (m *Bit) Clear(flag Bit)    { *m = *m &^ flag }
func (m *Bit) Toggle(flag Bit)   { *m = *m ^ flag }
func (m *Bit) Has(flag Bit) bool { return *m&flag != 0 }
