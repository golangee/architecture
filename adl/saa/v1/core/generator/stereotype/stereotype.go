package stereotype

type secretKey string

const (
	// e.g. by bash SET X=my_var. Value is either nil or a string.
	kEnvironmentVariable secretKey = "kEnvironmentVariable"

	// e.g. by calling ./myprogram -X=my_var. Value is either nil or a string.
	kProgramFlagVariable secretKey = "kProgramFlagVariable"

	// denotes if a struct is used for external configuration (env and program flags). Either nil, true or false.
	kConfiguration secretKey = "kConfiguration"

	// denotes if a struct is used for database configuration (env and program flags). Either nil, true or false.
	kDBConfiguration secretKey = "kDBConfiguration"

	// denotes if is mysql related.
	kMySQLRelated secretKey = "kMySQLRelated"

	// denotes a mysql table name
	kSQLTableName secretKey = "kSQLTableName"

	// denotes a mysql column name
	kSQLColumnName secretKey = "kSQLColumnName"
)
