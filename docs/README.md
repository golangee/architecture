# architecture

## The model

![](Architecture_Model.png)

TODO This section is currently just a collection of thoughts, it needs to be rewritten to have a coherent structure. Files for this experiment can be found in `arc/model` and `testdata/supportiety` but might not be up to date.

**Service** Services are either of type *core* or *io*. Core services contain the business logic while io services communicate with the outside world, via a user interface, network, hard drive, etc. To enforce a clean architecture the following rules of communication between services are established:

 * core <-> core is allowed
 * core <-> io is allowed
 * io <-> world is allowed
 * core <-> world is not allowed
 * io <-> io is not allowed

This separation will make sure that data concerning the domain must flow through core services. Ideally we would prevent core services from accessing the network or disk, but this might be hard to enforce in code. Still, programmers should respect this model and we can check for service dependencies that violate these rules in the generator.

To enable communication, each service provides a list of *ServiceDependencies*. These are the names of services it needs access to. When generating the code, the dependency injection layer will provide references to each required service, while not violating the above mentioned communication rules. Dependencies on external libraries can be specified in *LibDependencies*. In this array generator independent dependencies should be specified. E.g. "xml" could be used if the service needs access to an XML-Encoder, or "request" if it needs to make HTTP request. The realizations for these dependencies are given in the generators (DependencyDefinition::Id).

**Domain** The domain describes your whole business domain. It has name, a description and an *ArcVersion* to describe which version of this framework it uses. It consists of several things: multiple BoundedContexts, Glossaries and Executables.

**Executable** An executable is one that can be compiled and run. To configure an executable you can select from several generators using the *GeneratorSelection*. The functionality of an executable is composed of several services.

TODO What happens when services that depend on each other are defined in multiple services? Is that allowed and some communication layer is generated?

**Glossary** A glossary contains all explanations for items that are relevant in your domain. There is a glossary for your domain containing all items valid accross all BoundedContexts as well as a glossary for each BoundedContext that can contain items that are only valid in this context or override definitions from the domain. Several things, like the role in stories or the view for a DTO *must* be defined in the glossary to make sure that you are building a well defined system.

**BoundedContext** A bounded context defines a solution space for a problem in your domain. It has a name and a description to further explain it. A BoundedContext is built by several authors who have a name and a mail address to contact them. A BoundedContext also has a license that is one of the [SPDX license identifiers](https://spdx.org/licenses/). A BoundedContext describes a solution that is composed of several loosely coupled *packages*. It can have arbitrarily many packages but always has a standard package called *core* where all services without an explicitly defined package go.

**Package** // TODO //

**Method** Each service is put together from several methods. Each method can have several parameters and several results, that must be DTOs or Errors, since only these can be exchanged between services. It also has a linked *Story* that is the id of a story that justifies the existence of this method. Each method must have an appropriate story associated with it, otherwise it is not valid.

**Story** User-Stories are used to describe the functionality of the system. Since each service method *requires* a story, they are the backbone of building an application of any kind. Stories have a title and their content follows the form "As a *role* I want *functionality* so that *benefit*." The properties are named so that you can read a story directly:

```
#as_a      User
#i_want_to create new support tickets
#so_that   I get help from the support team
```

**AcceptCriterion** Stories must also have at least one AcceptCriterion. Should a tester find all criteria sufficiently fulfilled, then the story can be seen as completed. Each criterion has several *require*ments, that define conditions that must hold true, and several triggers that define *when* some expected action is to take place, to *then* give some results. There needs to be at least on item in *require*, *when*, *then* for this to be a valid AcceptCriterion.

**GeneratorSelection** allows an executable to choose what kind of project(s) should be generated for it. In the diagram a *GoGenerator* and an *AndroidGenerator* are shown. The AndroidGenerator is only there to show that different generators could be implemented. In the following only the more complete GoGenerator will be explained. The GoGenerator specifies a package in which the project should be generated into, a list of target platforms to build for and a list of dependencies.

**DependencyDefinition** Each definition consists of an *Id*, *Type*, *Name* and *Version*. Name and version will be used to generate a usable dependency. The Id is the name that services use to reference it. E.g. an id `xml` could be used to reference `encoding/xml` version `1.17`. *Type* should be set appropriately, so that dependencies that communicate with the outside world (like network or disk) are marked `io` and dependencies that do not communicate (like an xml encoder) are marked `core`. This distinction is useful to enforce the communication rules for keeping a clean architecture in section *Service*. Core-services can only have core dependencies and io-services can have io and core dependencies.

**DTO** Data Transfer Objects are used excessively in this project. They are the only things that can be communicated between services. Whenever some data needs to be transferred, a DTO must be used (hence the name). In this project DTOs are views on an abstract type (specified in `Viewing` in the diagram). This type must be defined in the glossary for the bounded context (or the domain). A DTO itself is then a manifestation of that type with certain attributes that are only needed in a certain service. E.g. a `User` is described in the glossary. One service might be used for creating users and expects a `RegisterUser(name, email, password)` DTO. It will do some validation and then send a `CreateUser(name, email, password, id)` DTO to the database service to persist the user. When someone wants to look at this user, another service can send a `ShowUser(name)`. Notice how each DTO describes the same concept of a user, but from different views, which is why they would all have the attribute `Viewing="User"` set. This is helpful since e.g. the `ShowUser` DTO should probably not send email and password for a user into the world.

**Error** Not all operations might complete successfully which is why the concept of errors is important. An error is an enum with several variants. A *ReadUserError* could have a *FileNotFound* and a *InvalidFileFormat* variant.