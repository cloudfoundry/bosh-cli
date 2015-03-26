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
    admin:
      client_secret: PASSWORD
    client:
      autoapprove:
      - bosh_cli
    clients:
      bosh_cli:
        id: bosh_cli
        override: true
        authorized-grant-types: implicit,password,refresh_token
        scope: openid
        authorities: uaa.none
        secret: ""
    cc:
      token_secret: "uaa-secret-key"
    scim:
      users:
      - marissa|koala|marissa@test.org|Marissa|Bloggs|uaa.user
      userids_enabled: true
    url: https://ADDRESS
    login:
      client_secret: PASSWORD
    ssl:
      key: SSL_CERTIFICATE_KEY
      cert: SSL_CERTIFICATE

  domain: example.com

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

### Notes
* uaa.nginx_port must be 443 due to tomcat redirect which ignores forwarded port
