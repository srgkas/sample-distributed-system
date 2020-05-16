package model

type Sensor struct {
	ID int `json:"id"`
	Name string `json:"name"`
}

func GetSensorByName(name string) (Sensor, error) {
	row := connection.QueryRow("SELECT id, name FROM sensors WHERE name=$1", name)

	sensor := Sensor{}

	err := row.Scan(&sensor.ID, &sensor.Name)

	return sensor, err
}