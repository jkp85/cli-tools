[![slack in](https://slackin-pypmyuhqds.now.sh/badge.svg)](https://slackin-pypmyuhqds.now.sh/)

# Command Line Interface (CLI) Tools

Command Line Interface (CLI) tools used to manage 3Blades resources. CLI tools connect to [3Blades backend API](https://github.com/3blades/app-backend).

## Local compilation

Install Go as instructed here [https://golang.org/doc/install](https://golang.org/doc/install).

Make sure you have go available in your shell.

	go version

Check if you have GOPATH in your env variables (default is $HOME/go)

It's good idea to add $GOPATH/bin folder to your PATH variable.

Get all you need:

	go get github.com/3Blades/cli-tools

If you added your bin folder from GOPATH to your PATH then you can just run:

	cli-tools

If don't then:

	cd $GOPATH/bin
	./cli-tools

If you have no error message then you are good to go :)

If you want to recompile it then:

	cd $GOPATH/src/github.com/3Blades/cli-tools
	go install


To update cli-tools you can do:

	go get -u github.com/3Blades/cli-tools

## Config

In order for cli-tools to work with your api server you need to put your api endpoint to config file.
CLI are looking for config file in your home directory. Default config file can be json, yaml or toml for example

	.threeblades.yaml

Currently supported options are:

	root: localhost:5000 // api root

## Workflow

An example will use tensorflow and keras for modelling.

**Note 1:** You need to set your apidomain or if you do local development to your machine IP here:
[http://localhost:5000/admin/sites/site/](http://localhost:5000/admin/sites/site/) It needs to be set with port.
For me it is `192.168.0.100:5000`.

**Note 2:** You need to create a resources instance for your server to run. You can do it here:
`http://localhost:5000/<username>/servers/options/resources/`. Keep your created resource id, because we will use it later.

**Note 3:** You will need your api token in order to make requests to model server. After you login to api with this cli tools, you can find your token inside a file `.threeblades.token` in the same directory as your config file. 

### Notebook

Create notebook:

	tbs server create --name keras_cpu --image keras --resources <resources_uuid> --type jupyter

Start notebook:

	tbs server start --name keras_cpu

Go to `http://localhost:5000/server/<notebook_id>/jupyter/tree`.
We will use [this dataset](http://archive.ics.uci.edu/ml/machine-learning-databases/pima-indians-diabetes/pima-indians-diabetes.data).
Download it and save as `pima-indians-diabetes.csv`. You can upload it to notebook with notebook ui on notebook home page.
We will create our model now:

```python
from keras.models import Sequential
from keras.layers import Dense
import numpy
import os

# fix random seed for reproducibility
numpy.random.seed(7)

# split into input (X) and output (Y) variables
X = dataset[:,0:8]
Y = dataset[:,8]

# create model
model = Sequential()
model.add(Dense(12, input_dim=8, kernel_initializer='uniform', activation='relu'))
model.add(Dense(8, kernel_initializer='uniform', activation='relu'))
model.add(Dense(1, kernel_initializer='uniform', activation='sigmoid'))

# Compile model
model.compile(loss='binary_crossentropy', optimizer='adam', metrics=['accuracy'])

# Fit the model
model.fit(X, Y, epochs=150, batch_size=10, verbose=0)

# evaluate the model
scores = model.evaluate(X, Y, verbose=0)
print("%s: %.2f%%" % (model.metrics_names[1], scores[1]*100))

# serialize model to JSON
model_json = model.to_json()
with open("model.json", "w") as json_file:
    json_file.write(model_json)
print("Saved model to disk")
```

In order to use our new saved model as restful endpoint we need to create a script that will load our model
and process data.

### Restful model

In notebook create a file called `model.py` and paste following code:

```python
import json
import numpy

from keras.models import model_from_json


def load_model():
    with open('model.json', 'r') as json_file:
        loaded_model_json = json_file.read()
    loaded_model = model_from_json(loaded_model_json)
    return loaded_model


def main(json_data):
	"""
    Example json_data:
	{
        "X": [[6,148,72,35,0,33.6,0.627,50], [1,85,66,29,0,26.6,0.351,31]]
	}
	"""
    data = json.loads(json_data)
    model = load_model()
    X = numpy.asarray(data["X"])
    predictions = model.predict(X)
    return json.dumps(predictions.tolist())
```

Save the file.

Go to console and create a model server:

	tbs server create --name keras_model --image keras --resources <resources_uuid> --type restful --script model.py --function main

Start your model server:

	tbs server start --name keras_model

Send a request to your new model server:

```
curl -X POST -d '{"schema_version": "0.1", "model_version":"1.0", "timestamp":"2017-04-04T17:43:45.022569696Z", "data": {"X": [[6,148,72,35,0,33.6,0.627,50], [1,85,66,29,0,26.6,0.351,31]]}}' -H "Authorization: Token <your_token>" http://localhost:5000/server/<server_uuid>/restful/
```

You should get a response similar to this:

```
{"schema_version":"0.1","model_version":"1.0","timestamp":"2017-04-05T13:09:00.052713946Z","status":"ok","execution_time":62,"data":[[0.5022257566452026],[0.5014487504959106]]}
```
