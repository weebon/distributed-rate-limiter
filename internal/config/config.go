package config

 
type RouteLimit struct {
	Capacity   float64
	RefillRate float64  
	WindowSecs float64  
	Limit      int64    
}

 type Config struct {
	Routes  map[string]RouteLimit
	Default RouteLimit
}

 func NewDefaultConfig() *Config {
	return &Config{
		Default: RouteLimit{
			Capacity:   5,
			RefillRate: 1,
			WindowSecs: 10,
			Limit:      5,
		},
		Routes: map[string]RouteLimit{
			"/api/search": {
				Capacity:   20,
				RefillRate: 5,
				WindowSecs: 10,
				Limit:      20,
			},
			"/api/upload": {
				Capacity:   2,
				RefillRate: 0.1,
				WindowSecs: 60,
				Limit:      2,
			},
		},
	}
}

 func (c *Config) ForRoute(path string) RouteLimit {
	if rl, ok := c.Routes[path]; ok {
		return rl
	}
	return c.Default
}