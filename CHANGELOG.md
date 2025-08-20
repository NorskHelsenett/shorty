# Changelog

All notable changes to this project will be documented in this file.

Must include:

    ## [X.X.X] - YYYY-MM-DD 
    ### **Environment:** 
    ### **Description:** 
    ### **Impact:** 

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

----

## [1.0.0-rc67] - 2025-08-13

### **Environment:** Test/QA only
### **Description:** Performance improvements and bug fixes for URL shortening service
### **Impact:** Test teams, QA engineers

#### Added
- Enhanced URL validation with better error messages
- New metrics endpoint for monitoring URL creation rates
- Improved logging for debugging URL resolution issues

#### Changed  
- Updated Go runtime to 1.24-alpine for better performance
- Optimized database queries for faster URL lookups
- Enhanced Docker image size reduction (30% smaller)

#### Fixed
- Fixed URL expiration not working correctly in test environments
- Resolved memory leak in URL resolution service
- Fixed container startup race condition


---

## Template for future releases:

## [X.X.X] - YYYY-MM-DD

### **Environment:** 
### **Description:** 
### **Impact:** 

#### Added
- New features

#### Changed  
- Changes in existing functionality

#### Deprecated
- Soon-to-be removed features

#### Removed
- Removed features

#### Fixed
- Bug fixes

#### Security
- Security improvements
