# Security Policy

## Supported Versions

We currently support the following versions with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0.0 | :x:                |

## Dependency Management

We use Dependabot to automatically monitor and update our dependencies:

- Weekly dependency updates
- Automatic security vulnerability scanning
- Pull requests for dependency updates
- Manual review of all dependency changes
- Automatic assignment to security team for review

## Reporting a Vulnerability

We take the security of our library seriously. If you believe you have found a security vulnerability, please follow these steps:

1. **Do not disclose the vulnerability publicly**
2. Email our security team at security@kopexa.com with:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Any suggested fixes (if available)

### What to expect

- You will receive an acknowledgment within 48 hours
- We will investigate and keep you updated on our progress
- We will work with you to validate and address the vulnerability
- Once fixed, we will:
  - Credit you in the security advisory (unless you prefer to remain anonymous)
  - Release a patch as soon as possible
  - Update the CHANGELOG.md

## Security Best Practices

When using this library, we recommend:

1. Always use the latest stable version
2. Regularly update dependencies
3. Follow Go security best practices
4. Use security scanning tools in your CI/CD pipeline
5. Implement proper input validation
6. Use secure configuration management

## Security Updates

Security updates will be released as patch versions (e.g., 1.0.1) and will be clearly marked in the CHANGELOG.md. We recommend implementing a process to regularly check for and apply these updates. 