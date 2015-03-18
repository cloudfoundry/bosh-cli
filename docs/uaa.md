## Using UAA for authentication

### 1. Download CF Release.

Download a release from the releases page of [cf-release](http://bosh.io/releases/github.com/cloudfoundry/cf-release).

### 2. Add stuff to your manifest.

Add the `uaa` job to your deployment jobs:

    - { name: uaa, release: cf }

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
    jwt:
      signing_key: |
        -----BEGIN RSA PRIVATE KEY-----
        MIICXAIBAAKBgQDHFr+KICms+tuT1OXJwhCUmR2dKVy7psa8xzElSyzqx7oJyfJ1
        JZyOzToj9T5SfTIq396agbHJWVfYphNahvZ/7uMXqHxf+ZH9BL1gk9Y6kCnbM5R6
        0gfwjyW1/dQPjOzn9N394zd2FJoFHwdq9Qs0wBugspULZVNRxq7veq/fzwIDAQAB
        AoGBAJ8dRTQFhIllbHx4GLbpTQsWXJ6w4hZvskJKCLM/o8R4n+0W45pQ1xEiYKdA
        Z/DRcnjltylRImBD8XuLL8iYOQSZXNMb1h3g5/UGbUXLmCgQLOUUlnYt34QOQm+0
        KvUqfMSFBbKMsYBAoQmNdTHBaz3dZa8ON9hh/f5TT8u0OWNRAkEA5opzsIXv+52J
        duc1VGyX3SwlxiE2dStW8wZqGiuLH142n6MKnkLU4ctNLiclw6BZePXFZYIK+AkE
        xQ+k16je5QJBAN0TIKMPWIbbHVr5rkdUqOyezlFFWYOwnMmw/BKa1d3zp54VP/P8
        +5aQ2d4sMoKEOfdWH7UqMe3FszfYFvSu5KMCQFMYeFaaEEP7Jn8rGzfQ5HQd44ek
        lQJqmq6CE2BXbY/i34FuvPcKU70HEEygY6Y9d8J3o6zQ0K9SYNu+pcXt4lkCQA3h
        jJQQe5uEGJTExqed7jllQ0khFJzLMx0K6tj0NeeIzAaGCQz13oo2sCdeGRHO4aDh
        HH6Qlq/6UOV5wP8+GAcCQFgRCcB+hrje8hfEEefHcFpyKH+5g1Eu1k0mLrxK2zd+
        4SlotYRHgPCEubokb2S1zfZDWIXW3HmggnGgM949TlY=
        -----END RSA PRIVATE KEY-----
      verification_key: |
        -----BEGIN PUBLIC KEY-----
        MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDHFr+KICms+tuT1OXJwhCUmR2d
        KVy7psa8xzElSyzqx7oJyfJ1JZyOzToj9T5SfTIq396agbHJWVfYphNahvZ/7uMX
        qHxf+ZH9BL1gk9Y6kCnbM5R60gfwjyW1/dQPjOzn9N394zd2FJoFHwdq9Qs0wBug
        spULZVNRxq7veq/fzwIDAQAB
        -----END PUBLIC KEY-----
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


