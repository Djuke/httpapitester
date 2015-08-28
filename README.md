# HTTP API Tester
[![](http://dockeri.co/image/djukevanhoof/httpapitester)](https://hub.docker.com/r/djukevanhoof/httpapitester/)
[![](https://badge.imagelayers.io/djukevanhoof/httpapitester:latest.svg)](https://imagelayers.io/?images=djukevanhoof/httpapitester:latest 'Get your own badge on imagelayers.io')

# Table of Contents
- [Introduction](#introduction)
- [Usage](#usage)
- [Test suite](#test-suite)
  - [Test](#test)
    - [JSON response schema validation](#json-response-schema-validation)
  - [Default test](#default-test)
  - [First tests](#first-tests)
  - [Includes](#includes)
    - [Tests file](#tests-file)
  - [Last tests](#last-tests)
- [Running a test suite](#running-a-test-suite)
- [References](#references)

# Introduction

A tool to test a HTTP API if no tests are written for a HTTP service. I recommend writting HTTP API test as part of your code. But if you come across some code that hasn't this might be useful.

This code is tested and build using [Go](http://golang.org) version 1.5


# Usage

```bash
./httpapitester [test suite file]
```

Take a look at [testsuite_example.json](testsuite_example.json) for an example test suite or see [Test suite](#test-suite) for how to write a test suite.

```bash
./httpapitester ./testsuite.json
```

# Test suite

A test suite is a json format which is the start for writing tests. All properties of a test suite are optional.

```json
{
  "default":{},
  "first":[],
  "includes":[],
  "last":[]
}
```

See [testsuite_example.json](testsuite_example.json) for an example.

## Test

This is the test's JSON format:

```json
{
	"label":"example",
	"request":{
		"method":"GET",
		"url":{
			"scheme":"http",
			"host":"example.com",
			"path":"/example",
			"rawQuery":"var1=value1&var2=value2",
			"fragment":"exameple"
		},
		"urlUserInfo":{
		  "user":"example_user",
		  "pass":"example_pass"
		},
		"tlsInsecureSkipverify":false,
		"noDefaultHeaders":true,
		"headers":[
		  {
		    "key":"Content-type",
		    "value":"plain/text",
		    "useFromJar":true
		  }
		],
		"bodyString":"data",
		"bodyJson":{
			"name":"http api tester",
			"urls":[
				"github.com"
			]
		}
	},
	"response":{
	  "status":"200 OK",
		"statusCode":200,
		"noDefaultHeaders":true,
		"headers":[
		  {
		    "key":"Content-type",
		    "value":"plain/text",
		    "putInJar":false,
		    "validate":true
		  }
		]
		"bodyCheck":true,
		"bodyString":"success",
		"bodyJsonSchema":{}
	},
	"useCookieJar":true,
	"noCookieJar":false,
	"printDebugOnFail":false,
	"printJsonIndented":false
}
```

Explanation of some properties (not declaring a property has the same result as leaving it empty):
- **label**: will be printed when a test fails
- **request**: a request is contructed from this
  - **method**: specifies the HTTP method (GET, POST, PUT, etc.)
  - **url**: a url is contructed from this, the basic authentication credentials must be set using the `urlUserInfo` property
    - **scheme**: should be HTTP or HTTPS
    - **host**: host or host:port
    - **rawQuery**: encoded query values, without '?'
    - **fragment**: fragment for references, without '#'
  - **urlUserInfo**: basic authentication credentials, will be added to the `url` property
  - **tlsInsecureSkipverify**: controls whether to verify the server's certificate chain and host name. If true, TLS accepts any certificate presented by the server and any host name in that certificate. In this mode, TLS is susceptible to man-in-the-middle attacks.
  - **noDefaultHeaders**: if true the default headers will not be prepended
  - **headers**: header which will be added to the request
    - **useFromJar**: response header values can be put in the headerJar and then be used in the request
  - **bodyString**: can contain any sort of data and preceeds above `bodyJson` when not empty
  - **bodyJson**: added for readability within the test file, and it can be printed with indentation when the test fails, leave empty if no body should be send
- **response**: contains values which will be tested, leave empty if nothing should be checked
  - **noDefaultHeaders**: if true the default headers will not be added
  - **headers**: a default header will not overwrite the existing header
    - **validate**: validate the response header value
    - **putInJar**: the header value will be put in the headerJar which can be used by requests
  - **bodyCheck**: if true the body will be checked
  - **bodyString**: preceeds above `bodyJsonSchema` and is only tested if `bodyCheck` is true
  - **bodyJsonSchema**: see [JSON response schema validation](#json-response-schema-validation) for more information, will only be tested if `bodyCheck` is true
- **useCookieJar**: if true the global cookie jar will be used in the request and is updated on receiving the response
- **noCookieJar**: if true no cookie jar is used even if `useCookieJar` is true, see [Default test](#default-test) for more info
- **printDebugOnFail**: if tue and a test fails debug info is provided, see [Running a test suite](#running-a-test-suite) for an example
- **printJsonIndented**: if true and debug info is printed the request `bodyJson` and response body, if the response content type is `application/json`, will be printed with indentation for readability

### JSON response schema validation

This is a simple example of a JSON schema for validating the response body:

```json
{
  "label":"Example test",
  "request":{
    "method":"GET",
    "url":{
      "path":"/example"
    }
  },
  "response":{
    "statusCode":200,
    "bodyCheck":true,
    "bodyJsonSchema": {
      "title": "Example Schema",
      "type": "object",
      "properties": {
        "ID": {
          "description": "response code",
          "type": "integer",
          "minimum": 300
        },
        "name": {
          "type": "string"
        }
      },
      "required": ["ID", "name"]
    }
  }
}
```

See (http://json-schema.org) for more information on json schema.

## Default test

The default test describes which values to use in a [test](#test), [first](#first-tests) and [last](#last-tests) tests included, when none or, in some cases, false is provided.

Overwrite explanation
- **request**
  - **method**: default overwrites if empty
  - **url**
    - **scheme**: default overwrites if empty
    - **host**: default overwrites if empty
    - **rawQuery**: default overwrites if empty
    - **fragment**: default overwrites if empty
  - **urlUserInfo**: default overwrites if `url.host` is overwritten
  - **tlsInsecureSkipverify**: default will overwrite if `url.host` is overwritten and the default value is true
  - **headers**: a default header will not overwrite an existing header
- **response**
  - **contentType**: default overwrites if empty
  - **headers**: a default header will not overwrite an existing header
- **useCookieJar**: default overwrites if the default value is true
- **noCookieJar**: can not be overwritten by default and preceeds above `useCookieJar`
- **printDebugOnFail**: default overwrites if the default value is true
- **printJsonIndented**: default overwrites if `printDebugOnFail` is overwritten and the default value is true

## First tests

The `first` property can hold zero or more [tests](#test) which will be executed before [includes](#includes) and [last](#last-tests). 

## Includes

The `includes` property is a list of relative filepaths and directories. Filepaths should point to [tests files](#tests-file). A directory will be read recursively and should contain test files only or directories which contain no files or test files.

Also it's possible to write an `includes.json` which holds a includes list like the `includes` property of a test suite, the same rules apply. The tests are added in order of the list. For example:

```json
[
  "moretests.json"
  "folderwithtests"
  "folderwithmanytests/onlythisone.json"
  "folderwhichcontains/includes.json"
]
```

__NOTE__: When a directory is being read it will first look for a `includes.json` file if found it will stop reading the directory, instead the includes file will be read. If no `includes.json` file is found it will read the directories first and files second all in alphabetical order.

### Tests file

A tests file can hold zero or more tests, for example:

```json
[
  {
    "label":"Example test 1",
    "request":{
      "method":"GET",
      "url":{
        "path":"/example1"
      }
    },
    "response":{
      "statusCode":200,
      "bodyCheck":false
    }
  },
  {
    "label":"Example test 2",
    "request":{
      "method":"POST",
      "url":{
        "path":"/example2"
      }
    },
    "response":{
      "statusCode":200,
      "bodyCheck":false
    }
  }
]
```

## Last tests
The `last` property can hold zero or more [tests](#test) which will be executed after [first](#first-tests) and [includes](#includes).

# Running a test suite

If all went well you should see something like this:

```bash
Executed 6 of 6 (666.000000ms)
```

If one of the first tests failed:

```bash
Executed 1 of 6
FAILED second example in first of test suite
  expect status code to equal 200, given 418
Executed 2 of 6 (1 FAILED) (34.150981ms)
one of the first tests failed I will not continue to execute the other tests
exit status 1
```

If one or more tests other than first tests fail:

```bash
Executed 3 of 6 (661.975459ms)
FAILED first example
  expect status code to equal 201, given 200
Executed 4 of 11 (1 FAILED) (785.892161ms)
```

If the `printDebugOnFail` property is set to true, see [Test](#test), you should see something like this: 

```bash
Executed 3 of 6
FAILED first example
  expect status code to equal 200, given 418
DEBUG REQUEST
  URL: https://example.com/example1
  Headers: map[Accept-Charset:[utf-8] Content-Type:[application/x-www-form-urlencoded]]
  Body: var1=value1&var2=value2
DEBUG RESPONSE
  Headers: map[Expires:[Thu, 19 Nov 1981 08:52:00 GMT] Cache-Control:[no-store, no-cache, must-revalidate, post-check=0, pre-check=0] Pragma:[no-cache]]
  Status code: 418
  Status: 418 I'm a teapot
  Body: 
Executed 4 of 6
FAILED second example
  expect status code to equal 200, given 418
DEBUG REQUEST
  URL: https://example.com/example2
  Headers: map[Accept-Charset:[utf-8] Content-Type:[application/x-www-form-urlencoded]]
  Body:
DEBUG RESPONSE
  Headers: map[Expires:[Thu, 19 Nov 1981 08:52:00 GMT] Cache-Control:[no-store, no-cache, must-revalidate, post-check=0, pre-check=0] Pragma:[no-cache]]
  Status code: 418
  Status: 418 I'm a teapot
  Body: 
Executed 6 of 6 (2 FAILED) (34.150981ms)
```

# References
  * https://github.com/xeipuuv/gojsonpointer
  * https://github.com/xeipuuv/gojsonreference
  * https://github.com/xeipuuv/gojsonschema
  * http://json-schema.org
