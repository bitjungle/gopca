# David Thompson - 2014 - Nonlinear Dimensionality Reduction: KPCA
* https://www.youtube.com/watch?v=HbDHohXPLnU

JPL-Caltech Virtual Summer School Big Data Analytics
I'm David Thompson and this is the next
lecture in this series from the JPL
Caltech virtual summer school on big
data analytics
we've talked previously about different
strategies for dimensionality reduction
including basic feature selection
approaches linear methods like principal
component analysis and also metric
learning approaches which generalize
linear dimensionality reduction to the
case where a class information is
available now we're going to depart from
the realm of linear projections
altogether and move into a topic known
as non linear dimensionality reduction
and in particular we'll be describing
kernel principal component analysis
which is a kernel eyes diversion of PCA
all right the objectives of this talk
are to first just to be able to
distinguish linear from nonlinear
dimensionality reduction and to know and
understand the differences between them
kernel pca is just one of many different
nonlinear dimensionality reduction
strategies I think it's a particularly
useful and widespread one so we're gonna
talk about that and focus on that in
this talk and you should be comfortable
with it as and its derivation and also
know have some familiarity with the
other methods that are out there and
available to you if you'd prefer to use
those for other applications more
generally there are advantages and
pitfalls to nonlinear as opposed to
linear approaches and we should
understand try and understand those -
all right so why do we care about
 From subspaces to nonlinear manifolds
nonlinear projections well there are
some cases where linear subspaces just
aren't good enough I mentioned before
that linear subspaces are almost always
useful as on any data set but they may
not adequately represent the the
underlying manifold which may have some
other nonlinear structure right so even
if I start and I'm able to project my
data linearly down to say twenty or
forty dimensions that still might I
still might be able to find a more
efficient representation that captures
that models the data even better using a
lower dimensional manifold and that's
the idea behind nonlinear dimensionality
reduction so here's a an example here
portrayed in these four image panes we
have in the upper right our website yeah
the upper left on this s curve which is
a curved manifold as a two-dimensional
it's embedded in a three-dimensional
space now if you were to do a linear
dimensionality reduction on this data
set that would amount to maybe slicing
I'd taking a hyperplane somewhere
through that data and projecting all the
points onto that hyperplane they try as
you might you're not going to be able to
find a single hyperplane that would let
you recover the original points of that
of those data no matter how you orient
that hyperplane you're always going to
lose quite a bit of information and this
is reflected by the the PCA projection
result that you see there in the in the
upper right where it tends to mix a lot
of unlike colors together even though
the different colors are supposed to be
on different sides of the the manifold
so the PCI projection down to 2 D is
really inadequate even though this is
this structure is intrinsically
two-dimensional we can't capture that
with a linear projection the other two
panels in the bottom showed different
nonlinear dimensionality reduction
methods and they actually are capable of
modeling and unfolding this manifold too
in order to represent the data
adequately so and that's what nonlinear
dimensionality reduction strategies are
all about so some other examples of that
 Other examples
here is a case using image data we have
just two axes of variation here we're
looking at a face from different angles
and so the left-right pose is one axis
of variation and the up-down poses
another note that any two of these data
points that are close to each other in
this manifold are also going to be
similar to each other so this captures
sort of the local structure around all
data points and it also captures the
global structure that there are these
two main axes of variation but it's a
highly nonlinear relationship naturally
and if you were to try and embed this in
the space of pixel attributes you'd get
some highly folded structure probably
even though that the manifold itself is
continuous right because you could
continuously move this camera around and
get practically no change from between
neighboring data points so the local
structure is smooth and continuous but
of course it's a very complicated
manifold and this this is from early
work in in AI so map in order to
identify structures using nonlinear
dimensionality reduction here's another
result from the same paper this is
looking at handwritten digit recognition
so here we have lots of examples of twos
and you can imagine that these twos live
on a manifold of different different
manifestations of what a handwritten two
should look like right so if you change
any one of these just a little bit you
move along the manifold to its neighbors
right you're still on the manifold as
long as it's still valid - and this just
emphasizes the fact that these manifolds
needn't fill the space they aren't
always rectangular sheets you can have
spikes and pseudopods that jut out like
you see down there in the lower right
there's a sort of a spike in this
distribution representing this the twos
of the curly Q's on the top but that
said we can still sort of arbitrarily
apply ascribe different English titles
to these axes here they've been called
the bottom loop articulation and the top
arch articulation axes but really all
we're trying to do is describe come up
with some nonlinear structure to
describe the universe of possible twos
right which is a highly nonlinear
manifold in the original space how do we
represent this structure well this one
 Representing nonlinear structure Projecting to higher dimensions can simplify data that is not linearly separable.
way we can do this there are lots of
different ways a lot of different
nonlinear dimensionality reduction
methods involve looking at graphs for
instance to find local neighborhoods on
graphs but kernel principle component
analysis which is the one we're going to
be exploring most deeply in this talk
relies on the intuition that many data
sets which are not linearly separable in
their original attributes can be made
linearly separable by projecting them
into a higher dimensional space that is
we can add attributes which are simple
arithmetic operations of the original
attribute space which caused the data to
become linearly separable in that new
feature space so here's an example where
we have two data points this bulls-eye
data cloud has two classes the red and
the green which are not linearly
separable try as you might you cannot
draw separating discriminant line
through that or separating decision
boundary through that data cloud however
when you add a third feature to the data
set which is the sum of squared
attributes then the results becomes
linearly separable that you can see here
in this in this three-dimensional
portrayal so we're mapping this into a
high dimensional space where linear
relation
chips are sufficient if you were to map
that decision boundary that you make in
this higher dimensional space back into
the lower dimensional space you'd get a
nonlinear decision boundary so we can
actually use all of the tools from our
linear analysis toolkit in the high
dimensional space and get nonlinear
results in the original space this is an
important and important insight
important intuition is really the
foundation of kernel principle component
analysis so what does this look like for
K PCA kernel pca takes our training data
it Maps it into some higher dimensional
feature space where we perform principal
component analysis right which in in the
original data space it would be
represented as a nonlinear projection
right nonlinear transformation of the
data if we get some new data point then
we can similarly projected into our
higher dimensional feature space and
find its low D representation in that
new high dimensional feature space
that's that's through with projections
that have been learned from our training
data right so this is the intuition
behind K PCA but you might say well this
is this is fine but it's kind of
contrary to our the whole point of
dimensionality reduction right if we're
creating these arbitrary new high
dimensional feature spaces doesn't that
actually increase or than reduce the
dimensions it's true it would except on
we also have the benefit that for K PCA
you don't actually have to do any
calculations in the high dimensional
feature space and I'll describe that in
a moment
the derivation of kernel pca follows our
 Kernel PCA objective follows PCA
our PCA so we are now take some mapping
of the original data set which is fee so
fee of X sub I is the the Maps data
point in this high dimensional space and
we envision some projection in this high
dimensional space that's your
represented by this set of orthonormal
basis vectors and the matrix you write
and the expression here are u times u
transpose times V of X sub I gives us
the the reconstruction of the data point
V of X sub I
after down projecting it through this
this linear basis right and so we
subtract the original data point in the
high dimensional feature space from its
reconstruction in that high dimensional
feature space and we get an error score
right so we're trying to minimize
reconstruct
error in this high dimensional space
note that for this expression it's
actually important that the data set fee
be centered similarly to regular PCA we
were working with zero mean data we had
to subtract the mean off first here as
well we're going to be working with data
that's centered in the high dimensional
space alright so this is equivalent I
mentioned there previously the
relationship between singular value
decomposition and PCA so the same thing
applies here we can actually do a
singular value decomposition of the high
dimensional features are the high
dimensional data set V sub X and then
that gives us our our basis vectors U so
we define U to be the left singular
vectors of V sub X right associated with
the highest singular values in this
matrix of singular values Sigma all
right so equivalently right just as in
our previous pca example where there's
an equivalence between SVD and
calculating the eigen vectors of the
covariance matrix the sample covariance
matrix we can calculate the eigen
vectors of the sample covariance matrix
of V sub X which is a bunch of dot
products essentially so and again this
is this has to be lead the eigenvectors
of the the centered matrix as in in pca
okay this this brings me to the kernel
 The "kernel trick" Observation: PCA relies on eigenvectors of the sample covariance matrix XXT. This can be represented in the new feature space by the Kernel Matrix K, composed of dot products
trick so as before we can work with this
as before pca relies on the eigenvectors
of the sample covariance matrix x times
x transpose and we can represent that in
the projected space by a matrix of dot
products right and the expression for
that is given here so what this means is
that we can implicitly model any
nonlinear transformation feasts of of X
the only requirement is that we have to
be able to calculate dot products
between those data points efficiently we
can as long as we know the dot products
we can figure out what the projected
representation is so we actually never
have to calculate the explicit represent
high dimensional representation of the
new feature space so that typically the
 The kernel matrix K Contains all pairwise evaluations of the dot product
way this is done in kernel methods this
is known as the kernel trick is to come
up with a kernel matrix of dot products
between all the different data points
and so we described here the dot product
using a kernel fun
inque which represents points similarity
in this this high dimensional space or
similarity in the attribute space which
equates the their dot product right so
on this is different this is slightly
different than than PCA whereas PCA has
a covariance matrix that scaled with the
number of input dimensions n here we're
working in K PCA with this kernel matrix
that is has dimensions similar its
dimensions according to the size of our
dataset right because we're calculating
the dot product of all data points
against all the others so this scales
with D the number of data points in our
training set
what kernel function should we use how
 What kernel function to use? A typical choice is the Gaussian kernel
do we calculate these dot products well
a common method is to just look at a
local some sort of decreasing function
of distance in the original feature
space right as typical quois a typical
choice is the Gaussian kernel we
introduced this way back in the first
lecture right as a way to describe
locality between data points that has a
width parameter H that we can set using
cross-validation if we like and this
falls off rapidly as distance increases
so we can apply this to every two
pairwise - all the different data points
and figure out what their what their dot
products are in this high dimensional
space so we're ascribing a functional
form to their dot products which lets us
calculate this nonlinear projection okay
so there's a catch I mentioned before
that our data set has to be centered in
this high dimensional feature
representation that it has to be zero
mean and this won't be the case in for a
general kernel matrix that we derive
what this really means is that we we
want to come up with some feature some
I'm sorry some data points I'm projected
data point we'll call this V tilde of X
sub I which is the original fee of X sub
I minus the mean of all of the fees and
you can actually push this through some
algebra to figure out what its
implementing implications would be for
the kernel matrix as a whole so
different with the kernel matrix
elements defined as the dot products of
our fee of X sub I and X sub J we then
perform that substitution from the top
line and
it actually works out to some simple
arithmetic operations on the kernel
matrix that perform this centering
operation so sparingly I won't walk you
through the gory details but it
basically amounts to this operation that
we have to perform on the kernel matrix
after calculating the kernel function of
all data points in order to Center it
before we can do SVD alright so here's
 KPCA Recipe
the from beginning to end the the recipe
for kernel principal component analysis
first thing to do is pick a kernel
function I hardly recommend we start
with the Gaussian there are lots of
other kernel functions and a whole
literature devoted to defining different
kernel functions for different kinds of
input spaces so I encourage you to look
into that if it's a topic that interests
you after you have a kernel function you
can calculate the kernel matrix and then
Center it according to that expression
that I showed you before here I've
rented out a matrix notation with the
ones indicating just a matrix with with
entries of 1 over d where d is the
number of data points right so it's
fairly straightforward centering
operation and then you solve the the
eigen problem which is again this eigen
system based on the kernel matrix which
gives us our projections alpha and then
alpha is of course the size of all of
our data points in the training set so
in order to find the projection of a new
data point we multiply the alpha by its
kernel evaluation for all of the data
points in the training data with respect
to the query point that we're trying to
project and that gives us the location
along some new dimension for for the the
projected data point ok so here's an
example what it looks like in practice
so we've taken the face data set and
down projected it with PCA first so this
is a purely linear projection of that
that face data set before this is
courtesy code godsey in 2006
um you'll note that there are lots of
places where neighboring faces have
actually are neighboring points in this
data set have actually very different
images right so this this doesn't
necessarily do a good job of capturing
the underlying structure we can do
better with with KP CA now this isn't a
perfect unfolding of the manifold right
it doesn't totally fill the space with a
beautiful rectangle right but that's
not that rarely happens in practice
actually and this is actually a pretty a
pretty good result in that neighboring
data points also represent similar
images as you can see from the examples
that are plotted so this actually does a
much better job for this face data set
than a purely linear projection would as
long as we're working with just two
dimensions so kernel pca nonlinear
dimensionality reduction is a more
efficient method in this case
representing the data other methods for
 Other methods for nonlinear dimensionality reduction • Laplacian Eigenmaps • Local Linear Embedding
nonlinear dimensionality reduction and
indeed there are a bunch kernel pca is a
common one but there are lots of others
based on graph structure like I
mentioned laplacian eigen Maps or an
example of that it's a sort of local
linear embedding is another case I so
map is a famous algorithm that's been
around for over a decade now
multi-dimensional scaling is a classic
algorithm that takes a matrix of
affinities which in some ways similar to
the kernel matrix that I described
before and creates a low dimensional
projection that preserves those those
affinity values those distance values
feed-forward autoencoders are kind of an
exotic approach that's actually gained a
lot of traction of late in the deep
learning community so this is based on
neural network modeling and again that's
something that I encourage you to
investigate further
if neural networks are of interest
feed-forward autoencoders can be used to
perform nonlinear dimensionality
reduction irrespective of the ultimate
classification goal of the neural
network all right and of course there
there are many more strategies and a
huge literature on nonlinear
dimensionality reduction but I hope I've
given you a pretty good taste with this
lecture okay so one a note in closing
 Nonlinear dimensionality reduction works well when... • You have a lot of data (to fill the manifold) • The intrinsic dimensionality is relatively low . Your data is evenly distributed on the manifold
nonlinear dimensionality reduction
doesn't always work better than the
linear case it does work well when you
have a lot of data when you can fill the
manifold and you don't have a lot of
gaps or or outliers it works pretty well
in the intrinsic dimensionality is
relatively low and your data is pretty
evenly distributed on the manifold if
you don't have a lot of data then some
simpler structure is generally going to
give you better results occasionally
nonlinear dimensionality reduction
strategies can give you degenerate
solutions or they can be unstable and
and this is just goes with the territory
when you have a more flexible model
right you need more data in order to
fit it adequately so this it's just a
general caveat to keep in mind I would
always start with linear dimensionality
reduction strategies to see if they're
sufficient for your classification or
visualization task and only pull out
nonlinear dimensionality reduction when
it turns out to be absolutely necessary
it's particularly helpful in a lot of
cases like image data or you have smooth
transitions right from one frame to the
next or in pose estimation where you
have like say an articulated body that's
moving and can define relatively simple
low dimensional manifolds in that space
consisting of different poses and joint
limb articulations so there are some
sort of cottage industries where these
nonlinear representations have become
really useful but if you have some brand
new data set then start with the linear
okay so in summary methods like kernel
 Summary . Methods like KPCA can find non-linear structures by projecting data into high- dimensional spaces • You can avoid the cost of this calculation with the kernel trick, e.g. computations based on dot products
principle component analysis can find
nonlinear structures by projecting data
into higher dimensional spaces and then
applying the same linear toolkit from
PCA in those in those implicit spaces
and you can avoid the cost of this
calculation with the kernel trick that
is computations based on dot products
which are represented by a kernel
function that you can define