[![Codacy Badge](https://api.codacy.com/project/badge/Grade/960631824d444297870eb9d91dcafa2c)](https://www.codacy.com/app/3Blades/cli-tools?utm_source=github.com&utm_medium=referral&utm_content=3Blades/cli-tools&utm_campaign=badger)
[![Build Status](https://travis-ci.org/3Blades/cli-tools.svg?branch=master)](https://travis-ci.org/3Blades/cli-tools)
[![Build status](https://ci.appveyor.com/api/projects/status/im1ypes6cbo331xg?svg=true)](https://ci.appveyor.com/project/jgwerner/cli-tools)
[![slack in](https://slack.3blades.io/badge.svg)](https://slack.3blades.io)

# Command Line Interface (CLI) Tools

Command Line Interface (CLI) tools used to manage 3Blades resources. CLI tools connect to [3Blades backend API](https://github.com/3blades/app-backend).

Review our [online documentation](https://docs.3blades.io) for a full list of available command options.

## Local compilation

Install Go as instructed here [https://golang.org/doc/install](https://golang.org/doc/install). Although described in the installation instructions, the basic steps to set up Go are described below.

Although not obligatory, its a good idea to set your GOPATH working directory:

    mkdir $HOME/go
    GOPATH=$HOME/go

Make sure you have go available in your shell.

    go version

If you don't see any output, make sure your path reflects:

    PATH=$PATH:/usr/local/go/bin

Get all you need and compile:

    go get github.com/3Blades/cli-tools/tbs

If you added your bin folder from GOPATH to your PATH then you can just run:

    tbs

If don't then:

    cd $GOPATH/bin
    ./tbs

If you have no error message then you are good to go :)

If you want to recompile from local source then:

    cd $GOPATH/src/github.com/3Blades/cli-tools/tbs
    go install

To update cli-tools you can do:

    go get -u github.com/3Blades/cli-tools/tbs

## Config

In order for cli-tools to work with [3Blades API server](https://github.com/3blades/app-backend) you need to put your api endpoint to config file.
CLI are looking for config file in your home directory. Default config file can be json, yaml or toml for example

	$HOME/.threeblades.yaml

Currently supported options are:

	root: http://localhost:5000 // api root

## Workflow

An example will use tensorflow and keras for modelling.

**Note 1:** You need to set your apidomain or if you do local development to your machine IP here:
[http://localhost:5000/admin/sites/site/](http://localhost:5000/admin/sites/site/) It needs to be set with port.
For me it is `192.168.0.100:5000`.

**Note 2:** You will need your api token in order to make requests to model server. After you login to api with this cli tools, you can find your token inside a file `$HOME/.threeblades.token`.

### Notebook

Create notebook:

	tbs server create --name keras_cpu --image keras --type jupyter

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

## Bring your own node

In order to connect to your own node you need to install there docker engine and set it up to listen on a port.
After that you need run command:

	tbs host create --name HostName --ip "<your_host_ip>" --port 2375

Then you will be able to run your servers on your own node.

## Server logs

To stream server logs please use this command:

	tbs server logs --name <server_name>

Ctrl-C to interrupt stream.
