 #
    # Example Sieve Filter
    # Declare any optional features or extension used by the script
    #
    require ["fileinto"];

    #
    # Handle messages from known mailing lists
    # Move messages from IETF filter discussion list to filter mailbox
    #
    if header :is "Sender" "owner-ietf-mta-filters@imc.org"
            {
            fileinto "filter";  # move to "filter" mailbox
            }
    #
    # Keep all messages to or from people in my company
    #
    elsif address :DOMAIN :is ["From", "To"] "example.com"
            {
            keep;               # keep in "In" mailbox
            }

    #
    # Try and catch unsolicited email.  If a message is not to me,
    # or it contains a subject known to be spam, file it away.
    #
    elsif anyof (NOT address :all :contains
                   ["To", "Cc", "Bcc"] "me@example.com",
                 header :matches "subject"
                   ["*make*money*fast*", "*university*dipl*mas*"])
            {
            fileinto "spam";   # move to "spam" mailbox
            }
    else
            {
            # Move all other (non-company) mail to "personal"
            # mailbox.
            fileinto "personal";
            }

# NEW

require ["fileinto", "reject", "vacation", "regex", "relational", "comparator-i;ascii-numeric"];

if a :matches b {
  Do W; #an action
  stop;
}
elsif a :matches c {
  Do X;
  stop;
}
elsif a :matches d {
  Do Y;
  stop;
}
else {       # Nothing matches, put it into the Undecided folder and stop
  fileinto "INBOX.Undecided";
  stop;
}

require "virustest";
require "fileinto";
require "relational";
require "comparator-i;ascii-numeric";

/* Not scanned ? */
if virustest :value "eq" :comparator "i;ascii-numeric" "0" {
  fileinto "Unscanned";

/* Infected with high probability (value range in 1-5) */
} if virustest :value "eq" :comparator "i;ascii-numeric" "4" {
  /* Quarantine it in special folder (still somewhat dangerous) */
  fileinto "Quarantine";

/* Definitely infected */
} elsif virustest :value "eq" :comparator "i;ascii-numeric" "5" {
  /* Just get rid of it */
  discard;
}

require "spamtestplus";
require "fileinto";
require "relational";
require "comparator-i;ascii-numeric";

/* If the spamtest fails for some reason, e.g. spam header is missing, file
 * file it in a special folder.
 */
if spamtest :value "eq" :comparator "i;ascii-numeric" "0" {
  fileinto "Unclassified";

/* If the spamtest score (in the range 1-10) is larger than or equal to 3,
 * file it into the spam folder:
 */
} elsif spamtest :value "ge" :comparator "i;ascii-numeric" "3" {
  fileinto "Spam";

/* For more fine-grained score evaluation, the :percent tag can be used. The
 * following rule discards all messages with a percent score
 * (relative to maximum) of more than 85 %:
 */
} elsif spamtest :value "gt" :comparator "i;ascii-numeric" :percent "85" {
  discard;
}

/* Other messages get filed into INBOX */