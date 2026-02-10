package types

type GoAuth struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
}

type AddrConfig struct {
	UserSvcAddr string
}
