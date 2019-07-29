package client

type DeType struct {
	Name     string
	settings map[string]string
	port     int
}

func NewDeClient(m map[interface{}]interface{}) ClientType {
	var nc ClientType
	nc.Name = "Deluge"
	return nc
}

func (c DeType) Init() error {
	// resp, err := c.client.PostForm(c.settings["host"]+"/login", url.Values{
	// 	"username": {c.settings["username"]},
	// 	"password": {c.settings["password"]},
	// })
	// if err != nil {
	// 	log.Printf("Failed to initialize client: %v\n", err)
	// 	return err
	// }
	// resp.Body.Close()
	return nil
}

func (c DeType) Add(data []byte, filename string) error {
	// var b bytes.Buffer
	// w := multipart.NewWriter(&b)

	// // Write config.
	// for _, v := range qBparalist {
	// 	if c.settings[v] != "" {
	// 		w.WriteField(v, c.settings[v])
	// 	}
	// }
	// // Write torrent body.
	// h := make(textproto.MIMEHeader)
	// h.Set("Content-Disposition",
	// 	fmt.Sprintf(`form-data; name="torrents"; filename="%s"`, filename))
	// p, _ := w.CreatePart(h)
	// p.Write(data)
	// w.Close()

	// req, err := http.NewRequest("POST", c.settings["host"]+"/command/upload", &b)
	// if err != nil {
	// 	return err
	// }
	// // Don't forget to set the content type, this will contain the boundary.
	// req.Header.Set("Content-Type", w.FormDataContentType())

	// resp, err := c.client.Do(req)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }
	// body, _ := ioutil.ReadAll(resp.Body)
	// if string(body) != "Ok." {
	// 	return errors.New(string(body))
	// }

	// resp.Body.Close()
	return nil
}
