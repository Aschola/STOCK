# STOCK
This is a role based stock system with different users who have different roles


## Database Setup

1. **Create Database:**
   ```sh
   mysql -u root -p < db/create_database.sql

   ## Setting Up the Database

1. **Ensure Flyway is Installed**: Download and install Flyway from [Flyway's website](https://flywaydb.org/download/).

2. **Configure Flyway**:
   - Create a `flyway.conf` file in the root of your project directory.
   - Update the `flyway.conf` file with the following configuration:
     ```ini
     flyway.url=jdbc:mysql://localhost:3306/stock
     flyway.user=your_username
     flyway.password=your_password
     flyway.locations=filesystem:./migrations
     ```

3. **Apply Migrations**:
   - Run the following command to apply the migrations:
     ```bash
     flyway migrate
     ```

4. **Verify Setup**:
   - Ensure the database schema is up to date by checking the tables and schema.

