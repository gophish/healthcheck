# Gophish Healthcheck

A service to test mail servers for best practices.

### Current Status

This is considered **pre-alpha** at this point. Development is still _very much_ ongoing.

### Testing for Best Practices

When setting up a mail server, there's huge lists of best practices that describe what security settings to apply. Things like enforcing email authentication (SPF, DKIM, DMARC), filtering out certain attachment filetypes, adding external subject tags, and more.

Setting up the configuration can be tricky enough, but how do you test it? Not only that, but how do you _continously_ test it to make sure your security policy doesn't decrease over time? And then how do you test _every_ mail server?

That's the problem that Gophish Healthcheck aims to solve.