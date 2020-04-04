#include <stdio.h>
#include <stdlib.h>
#include <errno.h>
#include <sys/utsname.h>

extern char **environ;

int main(int argc, char *argv[]) {

   struct utsname buffer;

   errno = 0;
   if (uname(&buffer) != 0) {
      perror("uname");
      exit(EXIT_FAILURE);
   }

   // Print system information
   // http://man7.org/linux/man-pages/man2/uname.2.html
   printf("OS name: %s\n", buffer.sysname);
   printf("Hardware identifier: %s\n", buffer.machine);
   printf("\n");

   // Print arguments
   printf("Arguments:\n");
   for(int counter=0; counter<argc; counter++) {
		printf("argv[%d]: %s\n",counter,argv[counter]);
   }
   printf("\n");

   // Print environment
   printf("Environment:\n");
   char *s = *environ;
   for (int i = 1; s; i++) {
      printf("%s\n", s);
      s = *(environ+i);
   }
   printf("\n");

   int count = 0;
   while(1) {
      count++;
      printf("Waiting 10 seconds (%d)...\n", count);
      sleep(10);
   }

   return EXIT_SUCCESS;
}