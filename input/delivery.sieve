require ["fileinto", "envelope"];
if address :is "to" "dovecot@dovecot.org" {
  fileinto "Dovecot-list";
} elsif envelope :is "from" "owner-cipe-l@inka.de" {
  fileinto "lists.cipe";
} elsif anyof (header :contains "X-listname" "lugog@cip.rz.fh-offenburg.de",
               header :contains "List-Id" "Linux User Group Offenburg") {
  fileinto "ml.lugog";
} else {
  # The rest goes into INBOX
  # default is "implicit keep", we do it explicitly here
  keep;
}