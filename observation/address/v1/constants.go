package v1

// AccountType is the account type of an address across the observed chains.
type AccountType string

// Account types.
const (
	AccountTypeEOA      AccountType = "eoa"
	AccountTypeContract AccountType = "contract"
	AccountTypeBlank    AccountType = "blank"
)

// IsValid reports whether t is a known account type.
func (t AccountType) IsValid() bool {
	switch t {
	case AccountTypeEOA, AccountTypeContract, AccountTypeBlank:
		return true
	}
	return false
}

// Qualification is the qualification of an address on one chain.
type Qualification string

// Per-chain qualifications.
const (
	QualificationEOA         Qualification = "eoa"
	QualificationContract    Qualification = "contract"
	QualificationNotDeployed Qualification = "not_deployed"
	QualificationBlank       Qualification = "blank"
)

// IsValid reports whether q is a known per-chain qualification.
func (q Qualification) IsValid() bool {
	switch q {
	case QualificationEOA, QualificationContract, QualificationNotDeployed, QualificationBlank:
		return true
	}
	return false
}
