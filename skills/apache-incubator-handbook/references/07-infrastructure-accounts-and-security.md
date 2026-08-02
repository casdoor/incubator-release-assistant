# Infrastructure, accounts, records, and security

Last verified: 2026-08-02.

## Initial infrastructure

After IPMC acceptance, Mentors normally coordinate setup with ASF
Infrastructure. Typical sequence and resources include:

- podling metadata and incubation status page;
- LDAP/project group and `project.apache.org` DNS;
- public `dev@` and `commits@` lists plus private PPMC communication;
- optional `users@`, issues, pull-request, builds, or security lists;
- ASF-controlled Git/SVN repository and GitHub mirror/integration;
- issue tracker, CI, website publication, wiki, and distribution areas.

Use current ASF self-service or Infra request procedures. Do not assume an old
Jira request template or access mechanism still applies.

## Accounts and access

- A new committer is invited through the project's governance process before an
  ASF account is created.
- A new ASF committer needs an accepted ICLA. The ICLA grants the ASF sufficient
  contribution rights; it does not transfer the contributor's copyright.
- Account, project-group, repository, mailing-list, and GitHub access are
  separate systems. Confirm each rather than treating GitHub membership as the
  authoritative roster.
- Apache IDs and public signing fingerprints are public operational identifiers;
  passwords, recovery material, private keys, tokens, private-list content, and
  vulnerability details are not.

## Durable records

Preserve public evidence for:

- discussions, decisions, votes, and results;
- releases and verification;
- new committers/PPMC members after confidential voting is complete;
- reports and Mentor sign-off;
- status-page and roster changes;
- grants, IP clearance, and public legal decisions;
- graduation or retirement discussion and results.

Mailing-list archives are preferred for decisions because they remain visible
to participants across time zones and organizations. Summarize conclusions from
meetings or chat back to the public list.

## Security vulnerabilities

- Undisclosed vulnerabilities are reported privately to a project-specific
  `security@project.apache.org` address, or to `security@apache.org` when no
  project contact exists.
- Do not create a public GitHub/Jira issue or discuss details on a public list
  before coordinated disclosure.
- The project works privately with the reporter and ASF Security, assesses
  affected versions, prepares and reviews a fix, produces a release containing
  the fix, obtains CVE identifiers through the ASF process, then coordinates
  the advisory and public announcement.
- After disclosure, update the project's public security page and references.
- Support requests about already public vulnerabilities belong on normal user
  channels, not the private vulnerability-reporting address.

Security details are one of the legitimate exceptions to the default rule that
project work happens in public. Reveal only what is necessary to the authorized
response group until disclosure is coordinated.

## Operational hygiene

- Keep secrets and private-list material out of repositories, build logs,
  release evidence, and AI prompts/archives.
- Use project-controlled accounts and recoverable shared procedures rather than
  one person's credentials.
- Treat access changes, compromised keys/tokens, and inactive security response
  as issues requiring prompt escalation through the proper ASF channel.

Official sources:

- https://infra.apache.org/infra-incubator.html
- https://infra.apache.org/infra-contact.html
- https://infra.apache.org/new-committers-guide.html
- https://incubator.apache.org/guides/podling_sourcecontrol.html
- https://www.apache.org/security/
- https://www.apache.org/security/committers.html
