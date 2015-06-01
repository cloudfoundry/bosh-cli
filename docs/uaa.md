## Using UAA for authentication

### 1. Download the UAA BOSH Release.

Download a release from the releases page of [uaa-release](https://github.com/pivotal-cf-experimental/tmp-bosh-uaa-release/).

### 2. Add stuff to your manifest.

Add the `uaa` job to your deployment jobs:

    - { name: uaa, release: uaa }

Add uaa job properties to either the global properties or the job properties.

The following is an example property set:

```yaml
properties:
  director:
    user_management:
      provider: uaa
      options:
        key: uaa-secret-key
        url: https://ADDRESS
  uaa:
    db:
      address: DB-ADDRESS
      name: uaadb
      db_scheme: mysql
      port: 3306
      username: DB-USER
      password: DB-PASSWORD
    port: 25889
    nginx_port: 443
    admin: {client_secret: PASSWORD}
    client: {autoapprove: [bosh_cli]}
    clients:
      bosh_cli:
        id: bosh_cli
        override: true
        authorized-grant-types: implicit,password,refresh_token
        scope: openid
        authorities: uaa.none
        secret: ""
    cc: {token_secret: "uaa-secret-key"}
    scim:
      users:
      - marissa|koala|marissa@test.org|Marissa|Bloggs|uaa.user
      userids_enabled: true
    url: https://ADDRESS
    login: {client_secret: PASSWORD}
    ssl:
      key: SSL_CERTIFICATE_KEY
      cert: SSL_CERTIFICATE
  domain: example.com
  spring_profiles: mysql,default
  login:
    url: LOGIN_SERVER_URL
    entityBaseURL: LOGIN_SERVER_URL
    entityID: ENTITY_ID
```

To configure with LDAP add:

```yaml
properties:
	uaa:
    ldap:
      enabled: true
      profile_type: search-and-bind
      url: 'ldap://LDAP_HOST:389/'
      userDN: 'cn=admin,dc=test,dc=com'
      userPassword: 'password'
      searchBase: 'dc=test,dc=com'
      searchFilter: 'cn={0}'
```

to configure with SAML:

```yaml
properties:
  login:
    saml:
      serviceProviderKey:
      serviceProviderKeyPassword: password
      serviceProviderCertificate:
      nameID: 'urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified'
      assertionConsumerIndex: 0
      signMetaData: true
      signRequest: true
      socket:
        connectionManagerTimeout: 10000
        soTimeout: 10000
      providers:
        okta-local:
          idpMetadata: idpMetadata
          nameID: urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress
          assertionConsumerIndex: 0
          metadataTrustCheck: true
          showSamlLoginLink: true
          linkText: 'Okta Preview 1'
          iconUrl: 'http://link.to/icon.jpg'
```

to configure with client secret:

```yaml
properties:
  uaa:
    clients:
      test:
        id: test
        override: true
        authorized-grant-types: implicit,password,refresh_token,client_credentials
        scope: openid,password
        authorities: uaa.none
        secret: "secret"
```

### Notes

* uaa.nginx_port must be 443 due to Tomcat redirect which ignores forwarded port
* BOSH director is using UAA with symmetric key encryption. See [UAA docs](https://github.com/cloudfoundry/uaa/blob/master/docs/Sysadmin-Guide.rst) on how to configure UAA with symmetric key.
Currently UAA will be using symmetric key encryption if jwt:token:signing-key and jwt:token:verification-key are the same. Specifying cc:token_secret will render jwt token keys with the same value.
* See UAA logs in `/var/vcap/sys/log/uaa.log` in case of any issues.
* Make sure there is only one UAA service running if there are no logs.
* `spring_profiles` should specify database type that is used by UAA.
