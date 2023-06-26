require ["comparator-i;ascii-numeric","relational"];
if allof (
   not header :matches "x-spam-score" "-*",
   header :value "ge" :comparator "i;ascii-numeric" "x-spam-score" "10" )
{
  discard;
  stop;
}
