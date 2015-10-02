/*
Package xhub implements a simple web service for storing and retrieving the structured metadata of resources comprising an experimental research study.

The service recognizes three resource types (studies, trials, and files) and provides an API for the creation, listing, retrieval, and deletion of each type.

The primary unit of organization is the study. A study can contain trials or "study-level" files.  A trial can also contain files (viz., the ones associated with the experimental trial).  In other words, the xhub resource hierarchy aims to reflect the typical directory/file hierarchy of an experimental study:

    studies/

        STUDY_A/
            files/
                FILE_1
                FILE_2
                FILE_3
            trials/
                TRIAL_1/
                    FILE_3
                    FILE_4
                    FILE_5
                TRIAL_2/
                    FILE_6
                    FILE_7
                    FILE_8
                ...
                TRIAL_N/
                    ...

        STUDY_B/
            files/
                FILES
            trials/
                TRIAL/
                    FILES

        ...

        STUDY_N
            ...

Clients are expected to send resource representations via http POST requests with json-encoded payloads.  Clients can issue http GET requests for a list of resources (e.g., trials associated with a particular study) or a specific resource (e.g., a particular trial), where the http response will in turn be a json-encoded payload to be handled by the client.

---

TODO: Provide an overview of required/optional fields for incoming resource representations.  For now, see the Resource type used to handle POST requests.
*/
package xhub
