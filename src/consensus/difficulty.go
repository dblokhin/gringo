package consensus

// difficulty is defined as the maximum target divided by the block hash.
type Difficulty uint64

func (d Difficulty) Minimum() Difficulty {
	return Difficulty(MinimumDifficulty)
}

func (d Difficulty) FromNum(num uint64) Difficulty {
	return Difficulty(num)
}

func (d Difficulty) FromHash() Difficulty {
	/*
	pub fn from_hash(h: &Hash) -> Difficulty {
		let max_target = BigEndian::read_u64(&MAX_TARGET);
		// Use the first 64 bits of the given hash
		let mut in_vec = h.to_vec();
		in_vec.truncate(8);
		let num = BigEndian::read_u64(&in_vec);
		Difficulty { num: max_target / num }
	}

	*/
	return Difficulty(0)
}

func (d Difficulty) IntoNum() uint64 {
	return uint64(d)
}