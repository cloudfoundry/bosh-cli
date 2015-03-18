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
        url: http://ADDRESS:25888
  uaa:
    db:
      address: DB-ADDRESS
      name: uaadb
      db_scheme: mysql
      port: 3306
      username: DB-USER
      password: DB-PASSWORD
    port: 25888
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
        secret: uaa-secret-key
    cc:
      token_secret: "uaa-secret-key"
    scim:
      users:
      - marissa|koala|marissa@test.org|Marissa|Bloggs|uaa.user
      userids_enabled: true
    url: http://uaa.example.com
    login:
      client_secret: PASSWORD

  domain: example.com

  login:
    protocol: http
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
