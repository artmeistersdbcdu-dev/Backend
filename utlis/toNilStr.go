package utlis

import "database/sql"

func ToNilStr(str *string) sql.NullString {
	if str == nil {
		return sql.NullString{
			Valid: false,
		}
	}
	return sql.NullString{
		String: *str,
		Valid:  true,
	}
}
