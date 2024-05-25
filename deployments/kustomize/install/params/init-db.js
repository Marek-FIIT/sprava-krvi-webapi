const mongoHost = process.env.API_MONGODB_HOST
const mongoPort = process.env.API_MONGODB_PORT

const mongoUser = process.env.API_MONGODB_USERNAME
const mongoPassword = process.env.API_MONGODB_PASSWORD

const database = process.env.API_MONGODB_DATABASE
// const collection = process.env.API_MONGODB_COLLECTION

const retrySeconds = parseInt(process.env.RETRY_CONNECTION_SECONDS || "5") || 5;

// try to connect to mongoDB until it is not available
let connection;
while(true) {
    try {
        connection = Mongo(`mongodb://${mongoUser}:${mongoPassword}@${mongoHost}:${mongoPort}`);
        break;
    } catch (exception) {
        print(`Cannot connect to mongoDB: ${exception}`);
        print(`Will retry after ${retrySeconds} seconds`)
        sleep(retrySeconds * 1000);
    }
}

// if database and collection exists, exit with success - already initialized
const databases = connection.getDBNames()
if (databases.includes(database)) {
    const dbInstance = connection.getDB(database)
    collections = dbInstance.getCollectionNames()
    if (collections.includes("donor") && collections.includes("unit")) {
       print(`Collections 'donor' and 'unit' already exists in database '${database}'`)
        process.exit(0);
    }
}

// initialize
// create database and collection
const db = connection.getDB(database)
db.createCollection('donor')
db.createCollection('unit')

// create indexes
db['donor'].createIndex({ "id": 1 })
db['unit'].createIndex({ "id": 1 })

//insert sample data
let result1 = db['donor'].insertMany([
    {
        "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
        "birth_number": "9908121367",
        "first_name": "Peter",
        "last_name": "Marcin",
        "postal_code": "83407",
        "blood_type": "AB",
        "blood_rh": "+",
        "eligible": true,
        "last_donation": "2023-01-02T12:00:00Z",
        "email": "example.donor@mail.com",
        "phone_number": "+421905734825",
        "diseases": ["HIV", "Diabetes"],
        "medications": ["Paralen"],
        "substances": ["Alcohol", "Cocaine"],
        "created_at": "2023-01-01T12:00:00Z",
        "updated_at": "2023-01-02T12:00:00Z"
    },
    {
        "id": "6f47ac10b-58cc-4372-a567-0e02b2c3d478",
        "birth_number": "9908121377",
        "first_name": "Alice",
        "last_name": "Smith",
        "postal_code": "94025",
        "blood_type": "A",
        "blood_rh": "-",
        "eligible": false,
        "last_donation": "2023-02-01T10:00:00Z",
        "email": "alice.smith@example.com",
        "phone_number": "+421905734826",
        "diseases": ["Anemia"],
        "medications": ["Aspirin"],
        "substances": ["Nicotine"],
        "created_at": "2023-01-03T08:00:00Z",
        "updated_at": "2023-02-02T14:00:00Z"
    }  
]);

let result2 = db['unit'].insertMany([
    {
        "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
        "donor_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
        "donation_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
        "blood_type": "AB",
        "blood_rh": "+",
        "status": "available",
        "location": "83407",
        "contents": {
          "hemoglobin": 15.87,
          "erythrocytes": true,
          "leukocytes": true,
          "platelets": true,
          "plasma": true,
          "additional": ["alcohol"]
        },
        "frozen": false,
        "diseases": ["HIV", "Diabetes"],
        "expiration": "2023-01-01T12:00:00Z",
        "created_at": "2023-01-01T12:00:00Z",
        "updated_at": "2023-01-02T12:00:00Z"
    },
    {
        "id": "6f47ac10b-58cc-4372-a567-0e02b2c3d478",
        "donor_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
        "donation_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
        "blood_type": "AB",
        "blood_rh": "+",
        "status": "available",
        "location": "83407",
        "contents": {
          "hemoglobin": 15.87,
          "erythrocytes": true,
          "leukocytes": true,
          "platelets": true,
          "plasma": true,
          "additional": ["alcohol"]
        },
        "frozen": false,
        "diseases": ["HIV", "Diabetes"],
        "expiration": "2023-01-01T12:00:00Z",
        "created_at": "2023-01-01T12:00:00Z",
        "updated_at": "2023-01-02T12:00:00Z"
    }
]);

if (result1.writeError) {
    console.error(result1)
    print(`Error when writing the data: ${result1.errmsg}`)
}
if (result2.writeError) {
    console.error(result2)
    print(`Error when writing the data: ${result2.errmsg}`)
}

// exit with success
process.exit(0);