package main

var prepare = `Given the outline of the plot below, elaborating on the plot and
provide the characters in the story and locations of where the story plays out.
Give names to major characters and locations. Provide step by step progression 
of the story.

Outline:
--
%s
`

var first = `Write the first section of the story in detail, given the overall plot, 
setting the stage for the rest of the story. Keep the story open-ended
such that it can be easily continued in the next section. When possible, provide
the motivations of the various characters in the story.

Start with "# <title of section>"`

var next = `Continue fleshing the story and create the next section, following the 
overall plot and the story till now. Keep the section open-ended such that 
it can be easily continued in the next section. When possible, provide
the motivations of the various characters in the story.

Start with "# <title of section>".`

var last = `End the story with a twist and create the last section following the overall
plot and the story till now. When possible, provide the motivations of the various characters 
in the story. Wrap up the entire story as this is the last section. 

Start with "# <title of section>".`
