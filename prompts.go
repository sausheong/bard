package main

var prepare = `Given the outline of the plot below, elaborating on the plot and
provide the characters in the story and locations of where the story plays out.
Give names to major characters and locations. Provide step by step progression 
of the story.
--
%s
`

var first = `Write the first chapter of the story in detail, given the overall plot, 
setting the stage for the rest of the story. Keep the story open-ended
such that it can be easily continued in the next chapter. When possible, provide
the motivations of the various characters in the story.

Start with "# <title of chapter>"`

var next = `Continue fleshing the story and create the next chapter, following the 
overall plot and the story till now. Keep the chapter open-ended such that 
it can be easily continued in the next chapter. When possible, provide
the motivations of the various characters in the story.

Start with "# <title of chapter>".`

var last = `End the story with a twist and create the last chapter following the overall
plot and the story till now. When possible, provide the motivations of the various characters 
in the story. Wrap up the entire story as this is the last chapter. 

Start with "# <title of chapter>".`
