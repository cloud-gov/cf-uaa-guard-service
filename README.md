# UAA Auth Route Service [![Build Status](https://travis-ci.org/cloudfoundry-community/cf-uaa-guard-service.svg?branch=master)](https://travis-ci.org/cloudfoundry-community/cf-uaa-guard-service)

(Based on https://github.com/benlaplanche/cf-basic-auth-route-service)

Using the new route services functionality available in Cloud Foundry, you can now bind applications to routing services.
Traffic sent to your application is routed through the bound routing service before continuing onto your service.

This allows you to perform actions on the HTTP traffic, such as enforcing authentication, rate limiting or logging.

For more details see:
* [Route Services Documentation](http://docs.cloudfoundry.org/services/route-services.html)

## Getting Started

There are two components and thus steps to getting this up and running. The broker and the filtering proxy.

Before getting started you will need:

- Access to a cloud foundry deployment
- UAA client credentials

Uncomment and fill in the environment variables required as the sample in `broker-manifest.yml.sample` and copy the manifest to `broker-manifest.yml`.

Run `cf push -f broker-manifest.yml` to deploy the `uaa-guard-broker` app.

Uncomment and fill in the environment variables required as the sample in `proxy-manifest.yml.sample` and copy the manifest to `proxy-manifest.yml`.

Run `cf push -f proxy-manifest.yml` to deploy the `uaa-guard-proxy` app.

Once the broker is deployed, you can register it:

```sh
cf create-service-broker \
    uaa-auth-broker \
    $GUARD_BROKER_USERNAME \
    $GUARD_BROKER_PASSWORD \
    https://uaa-guard-broker.my-paas.com \
    --space-scoped
```

Once you've created the service broker, you must `enable-service-access` in
order to see it in the `marketplace`.

```sh
cf enable-service-access uaa-auth
```

You should now be able to see the service in the marketplace if you run `cf marketplace`

### Protecting an application with UAA authentication

Now you have setup the supporting components, you can now protect your application with auth!

First create an instance of the service from the marketplace, here we are calling our instance `authy`
```
$cf create-service uaa-auth uaa-auth authy
```

Next, identify the application and its URL which you wish to protect. Here we have an application called `hello` with a URL of `https://hello.my-paas.com`

Then you need to bind the service instance you created called `authy` to the `hello.my-paas.com` route
```
⇒  cf bind-route-service my-paas.com authy --hostname hello

Binding may cause requests for route hello.my-paas.com to be altered by service instance authy. Do you want to proceed?> y
Binding route hello.my-paas.com to service instance authy in org org / space space as admin...
OK
```

You can validate the route for `hello` is now bound to the `authy` service instance
```
⇒  cf routes
Getting routes for org org / space space as admin ...

space          host                domain            port   path   type   apps                service
space          hello               my-paas.com                            hello               authy
```

All of that looks good, so the last step is to validate we can no longer view the `hello` application without providing credentials!

```
⇒  curl -k https://hello.my-paas.com
Unauthorized
```

and if you visit it you will be redirected to UAA.

### Knowing who is logged in

This service will forward a header `X-AUTH-USER` with the email of the logged in user.
