
// - if push event in *preprocess.Container
// - - if payload contains .heupr.toml file
// - - - send request to GitHub using GetContents method
// - - - pull out Content field from results and decode base64
// - - - pass string(results) into the BurntSushi/toml library to get TOML struct
