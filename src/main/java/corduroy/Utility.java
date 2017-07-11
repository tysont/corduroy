package corduroy;

import java.net.InetAddress;
import java.net.UnknownHostException;
import java.nio.ByteBuffer;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;

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
}
