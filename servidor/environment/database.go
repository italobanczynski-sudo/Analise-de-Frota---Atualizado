package environment

type DatabaseConfig_t struct {
	URL string `envconfig:"DATABASE_URL" required:"true"`
}
