const express = require('express');
const bodyParser = require('body-parser')
var morgan = require('morgan')

const app = express();
const port = 9090;

app.use(bodyParser.json());       // to support JSON-encoded bodies
app.use(bodyParser.urlencoded({     // to support URL-encoded bodies
    extended: true
}));
app.use(morgan('combined'))

app.get('/', (req, res) => {
    console.log("This is a log")
    console.debug("This is a debug log")
    console.info("This is an info log")
    console.warn("This is a warn log")
    console.error("This is an error log")
    console.trace("This is a trace log")
    res.send({'active': true})
});

app.get('/healthz/', (req, res) => {
    res.send({'healthy': true})
});

app.get('/hello/', (req, res) => {
    res.send("Hello " + req.query?.name)
});

app.post('/proxy/', async (req, res) => {
    const {host, args} = req.body;
    const resp = await fetch(host.trim("/") + "/" + args.trim("/"))
    res.send(await resp.text())
});

app.listen(port, () => {
    console.log(`App listening on port ${port}`);
});
