// Author: Pirakalan

package definition

type CurrentWeather struct {
	Coord   *Coord     `json:"coord,omitempty"`
	Weather []*Weather `json:"weather,omitempty"`
	Base    string     `json:"-"`
	Main    *Main      `json:"main,omitempty"`
	Wind    *Wind      `json:"wind,omitempty"`
	Clouds  *Clouds    `json:"clouds,omitempty"`
	Rain    *Rain      `json:"rain,omitempty"`
	Snow    *Snow      `json:"snow,omitempty"`
	Dt      int        `json:"dt,omitempty"`
	Sys     *Sys       `json:"sys,omitempty"`
	CityId  int        `json:"id,omitempty"`
	Name    string     `json:"name,omitempty"`
	Code    int     `json:"cod"`
	Message string     `json:"message,omitempty"`
	UserStatus string `json:"-"`
}

type Coord struct {
	Lon float32 `json:"lon,omitempty"`
	Lat float32 `json:"lat,omitempty"`
}

type Weather struct {
	Id          int    `json:"id,omitempty"`
	Main        string `json:"main,omitempty"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

type Main struct {
	Temp     float32 `json:"temp,omitempty"`
	Pressure float32 `json:"pressure,omitempty"`
	Humidity float32 `json:"humidity,omitempty"`
	TempMin  float32 `json:"temp_min,omitempty"`
	TempMax  float32 `json:"temp_max,omitempty"`
}

type Wind struct {
	Speed float32 `json:"speed,omitempty"`
	Deg   float32 `json:"deg,omitempty"`
}

type Clouds struct {
	All float32 `json:"all,omitempty"`
}

type Rain struct {
	RainVolume3H float32 `json:"3h,omitempty"`
}

type Snow struct {
	SnowVolume3H float32 `json:"3h,omitempty"`
}

type Sys struct {
	Type    float32 `json:"-"`
	Id      int     `json:"-"`
	Message float32 `json:"-"`
	Country string  `json:"country"`
	Sunrise int     `json:"sunrise,omitempty"`
	Sunset  int     `json:"sunset,omitempty"`
}
