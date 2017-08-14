package corduroy;

import java.net.InetAddress;
import java.net.UnknownHostException;
import java.nio.ByteBuffer;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.*;

/**
 * Utility implements some generic functionality that can be used by other classes.
 */
public abstract class Utility  {

    /**
     * Gets a {@link String} that represents an address in the form 'host:port'.
     * @param inetAddress The host address.
     * @param port The port number.
     * @return The address.
     */
    public static String getAddress(InetAddress inetAddress, int port) {
        return inetAddress.getHostAddress() + ":" + port;
    }

    /**
     * Gets an {@link InetAddress} from a string address.
     * @param address The address.
     * @return The {@link InetAddress}.
     * @throws UnknownHostException
     */
    public static InetAddress getInetAddress(String address) throws UnknownHostException {
        return InetAddress.getByName(address.split(":")[0]);
    }

    /**
     * Gets a Port number from a string address.
     * @param address The address.
     * @return The Port number as an {@link int}.
     */
    public static int getPort(String address) {
        return Integer.parseInt(address.split(":")[1]);
    }

    /**
     * Concatenates two byte arrays.
     * @param a The first byte array.
     * @param b The second byte array.
     * @return The concatenated array.
     */
    public static byte[] concatenate(byte[] a, byte[] b) {
        byte[] c = new byte[a.length + b.length];
        System.arraycopy(a, 0, c, 0, a.length);
        System.arraycopy(b, 0, c, a.length, b.length);
        return c;
    }

    /**
     * Gets an {@link int} from a {@link byte} array.
     * @param b The byte array.
     * @return The int.
     */
    public static int byteArrayToInt(byte[] b) {
        ByteBuffer buffer = ByteBuffer.wrap(b);
        return buffer.getInt();
    }

    /**
     * Gets a {@link byte} array from an {@link int}.
     * @param i The int.
     * @return The byte array.
     */
    public static byte[] intToByteArray(int i) {
        ByteBuffer b = ByteBuffer.allocate(4);
        b.putInt(i);
        return b.array();
    }

    /**
     * Gets a positive integer hash representation of a {@link String} and a salt value.
     * @param s The string.
     * @param salt The salt value.
     * @return The hash.
     * @throws NoSuchAlgorithmException
     */
    public static int hash(String s, int salt) throws NoSuchAlgorithmException {
        return hash(s.getBytes(), intToByteArray(salt));
    }

    /**
     * Gets a positive integer hash representation of a {@link byte} array and a salt value.
     * @param b The byte array.
     * @param salt The salt value.
     * @return The hash.
     * @throws NoSuchAlgorithmException
     */
    public static int hash(byte[] b, byte[] salt) throws NoSuchAlgorithmException {
        byte[] b2 = concatenate(b, salt);
        MessageDigest md = MessageDigest.getInstance("SHA-1");
        return Math.abs(byteArrayToInt(md.digest(b2)));
    }

    /**
     * Gets the hash value of a node address with respect ot a specific ring number.
     * @param address The address of the node as a string.
     * @return The hash value as a positive integer.
     */
    public static int hashAddress(String address) {
        try {
            return hash(address, 1);
        }
        catch (Exception ex) {
            return -1;
        }
    }

    /**
     * Hashes a set of addresses.
     * @param addresses The addresses.
     * @return The hashes that represent the addresses.
     */
    public static Map<Integer, String> hashAddresses(Set<String> addresses) {

        Map<Integer, String> hashes = new HashMap<Integer, String>();
        for (String address : addresses) {
            int hash = hashAddress(address);
            hashes.put(hash, address);
        }

        return hashes;
    }

    /**
     * Finds the successor of a hash value in a sorted list of hash values.
     * @param hash The hash.
     * @param sortedHashes A list of hashes.
     * @return The first hash that is a successor to the given hash.
     */
    public static int findSuccessor(int hash, List<Integer> sortedHashes) {

        if (sortedHashes.isEmpty()) {
            return -1;
        }

        for (int sortedHash : sortedHashes) {
            if (sortedHash >= hash) {
                return sortedHash;
            }
        }

        return sortedHashes.get(0);
    }

    /**
     * Creates a finger table with entries that map to nodes in a ring.
     * @param address The address of the node the table is being created for.
     * @param addresses The addresses of other known nodes.
     * @return The finger table as a map of table entry numbers to addresses.
     */
    public static Map<Integer, String> createFingerTable(String address, Set<String> addresses) {

        int n = hashAddress(address);
        int m = Integer.MAX_VALUE;

        if (addresses.contains(address)) {
            addresses.remove(address);
        }

        Map<Integer, String> hashAddresses = hashAddresses(addresses);
        List<Integer> sortedHashes = new ArrayList(hashAddresses.keySet());
        Collections.sort(sortedHashes);

        Map<Integer, String> fingerTable = new HashMap<Integer, String>();
        for (int i = 1; i <= 4 * 8; i++) {

            int d = (int) Math.pow(2, i - 1);
            int v = ((n + d) % m);
            int h = findSuccessor(v, sortedHashes);
            String a = hashAddresses.get(h);

            if (!fingerTable.containsValue(a)) {
                fingerTable.put(i, a);
            }

        }

        return fingerTable;
    }
}
