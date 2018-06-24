# Gophish Healthcheck

A service to test mail servers for best practices.

### Current Status

This is considered **pre-alpha** at this point. Development is still _very much_ ongoing.

### Testing for Best Practices

When setting up a mail server, there's huge lists of best practices that describe what security settings to apply. Things like enforcing email authentication (SPF, DKIM, DMARC), filtering out certain attachment filetypes, adding external subject tags, and more.

Setting up the configuration can be tricky enough, but how do you test it? Not only that, but how do you _continously_ test it to make sure your security policy doesn't decrease over time? And then how do you test _every_ mail server?

That's the problem that Gophish Healthcheck aims to solve.

### What Does it Test?

To start, Gophish Healthcheck is going to let you send emails that either pass or fail email authentication. That is, we test:

* SPF
* DKIM
* DMARC

Once this is working, we'll add support for various attachment types, such as office documents with macros, executable files, and more.

If you have a different test you'd like to see added, [let us know!](https://github.com/gophish/healthcheck/issues)